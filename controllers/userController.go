package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"restaurant-management/database"
	"restaurant-management/helpers"
	"restaurant-management/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(ctx.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(ctx.Query("startIndex"))

		matchStage := bson.D{
			{Key: "$match", Value: bson.D{}},
		}
		projectStage := bson.D{{
			Key: "$project",
			Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{
					{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}},
				}},
			},
		}}

		result, err := userCollection.Aggregate(ctxWithTimeout, mongo.Pipeline{
			matchStage, projectStage,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
			return
		}

		var allUsers []bson.M
		if err = result.All(ctxWithTimeout, &allUsers); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allUsers)
	}
}

func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID := ctx.Param("user_id")
		var user models.User

		err := userCollection.FindOne(ctxWithTimeout, bson.M{"user_id": userID}).Decode(&user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
			return
		}
		ctx.JSON(http.StatusOK, user)
	}
}

func Signup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		// Convert the JSON data coming from postman to something that golang understands
		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate the data based of user struct
		validationErr := validate.Struct(user)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		// You'll check if the email has already been used by another user.
		count, err := userCollection.CountDocuments(ctxWithTimeout, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}

		if count > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "this email or phone number already exists"})
			return
		}

		// You'ill check if the phone number has already been used by another user
		count, err = userCollection.CountDocuments(ctxWithTimeout, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
			return
		}

		if count > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "this email or phone number already exists"})
			return
		}

		// Hash Password
		password := HashPassword(*user.Password)
		user.Password = &password

		// Create some extra details for the user object - created_at, updated_at, ID
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		// Generate tocken and refresh token (generate all tokens function from helper)
		token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)
		user.Token = &token
		user.Refresh_Token = &refreshToken

		// If all ok, then you insert this new user into the user collection
		resultInsertion, insertErr := userCollection.InsertOne(ctxWithTimeout, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		ctx.JSON(http.StatusOK, resultInsertion)
	}
}

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		var foundUser models.User

		// Convert the JSON data coming from postman to something that golang understands
		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Find a user with that email and see if that user even exists
		err := userCollection.FindOne(ctxWithTimeout, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		// Then you will verify the password
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if passwordIsValid != true {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		// If all goes well, then you'll generate tokens
		token, refreshToken, _ := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)

		// Update tokens - token and refresh
		helpers.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		// Return statusOK
		ctx.JSON(http.StatusOK, foundUser)
	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword, providePassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providePassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprint("login or password is incorrect")
		check = false
	}

	return check, msg
}
