package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"restaurant-management/database"
	"restaurant-management/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

func GetMenus() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		result, err := menuCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing the menu items"})
		}
		var allMenus []bson.M
		if err = result.All(ctxWithTimeout, &allMenus); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allMenus)
	}
}

func GetMenu() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		menuID := ctx.Param("menu_id")
		var menu models.Menu
		err := foodCollection.FindOne(ctxWithTimeout, bson.M{"menu_id": menuID}).Decode(&menu)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the menu item"})
			return
		}
		ctx.JSON(http.StatusOK, menu)
	}
}

func CreateMenu() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var menu models.Menu
		if err := ctx.BindJSON(&menu); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(menu)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		menu.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.ID = primitive.NewObjectID()
		menu.Menu_id = menu.ID.Hex()

		result, insertErr := menuCollection.InsertOne(ctxWithTimeout, menu)
		if insertErr != nil {
			msg := fmt.Sprintf("Menu item was not created")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		ctx.JSON(http.StatusOK, result)

	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}

func UpdateMenu() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var menu models.Menu
		if err := ctx.BindJSON(&menu); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		menuID := ctx.Param("menu_id")
		filter := bson.M{"menu_id": menuID}

		var updateObj primitive.D
		if menu.Start_Date != nil && menu.End_Date != nil {
			if !inTimeSpan(*menu.Start_Date, *menu.End_Date, time.Now()) {
				msg := "kindly retype the time"
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{"start_date", menu.Start_Date})
			updateObj = append(updateObj, bson.E{"end_date", menu.End_Date})

			if menu.Name != "" {
				updateObj = append(updateObj, bson.E{"name", menu.Name})
			}

			if menu.Name != "" {
				updateObj = append(updateObj, bson.E{"category", menu.Category})
			}

			menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			updateObj = append(updateObj, bson.E{"updated_at", menu.Updated_at})

			upsert := true
			opt := options.UpdateOptions{
				Upsert: &upsert,
			}

			result, err := menuCollection.UpdateOne(
				ctxWithTimeout,
				filter,
				bson.D{
					{"$set", updateObj},
				},
				&opt,
			)

			if err != nil {
				msg := "Menu update failed"
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			ctx.JSON(http.StatusOK, result)

		}
	}
}
