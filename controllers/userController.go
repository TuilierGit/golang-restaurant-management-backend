package controllers

import (
	"context"
	"log"
	"net/http"
	"restaurant-management/database"
	"restaurant-management/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

	}
}

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func HashPassword(password string) string {

}

func VerifyPassword(userPassword, providePassword string) (bool, string) {

}
