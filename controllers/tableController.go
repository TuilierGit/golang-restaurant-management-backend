package controllers

import (
	"context"
	"log"
	"net/http"
	"restaurant-management/database"
	"restaurant-management/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")

func GetTables() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := orderCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing table items"})
		}
		var allTables []bson.M
		if err = result.All(ctxWithTimeout, &allTables); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allTables)
	}
}

func GetTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		tableID := ctx.Param("table_id")
		var table models.Table
		err := tableCollection.FindOne(ctxWithTimeout, bson.M{"table_id": tableID}).Decode(&table)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the table item"})
			return
		}
		ctx.JSON(http.StatusOK, table)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var table models.Table
		if err := ctx.BindJSON(&table); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(table)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		table.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		table.ID = bson.NewObjectID()
		table.Table_id = table.ID.Hex()

		result, insertErr := tableCollection.InsertOne(ctxWithTimeout, table)
		if insertErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Table item was not created"})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var table models.Table
		tableID := ctx.Param("table_id")
		if err := ctx.BindJSON(&table); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj bson.D

		if table.Number_of_guests != nil {
			updateObj = append(updateObj, bson.E{
				Key:   "number_of_guests",
				Value: table.Number_of_guests,
			})
		}

		if table.Table_number != nil {
			updateObj = append(updateObj, bson.E{
				Key:   "table_number",
				Value: table.Table_number,
			})
		}

		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		upsert := true
		opt := options.UpdateOne().SetUpsert(upsert)

		filter := bson.M{"table_id": tableID}

		result, err := tableCollection.UpdateOne(
			ctxWithTimeout,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			opt,
		)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "table item update failed"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
