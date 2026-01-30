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

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

func GetOrders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := orderCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items"})
		}
		var allOrders []bson.M
		if err = result.All(ctxWithTimeout, &allOrders); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allOrders)
	}
}

func GetOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		orderID := ctx.Param("order_id")
		var order models.Order
		err := orderCollection.FindOne(ctxWithTimeout, bson.M{"order_id": orderID}).Decode(&order)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the order item"})
			return
		}
		ctx.JSON(http.StatusOK, order)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var table models.Table
		var order models.Order

		if err := ctx.BindJSON(&order); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(order)

		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		if order.Table_id != nil {
			err := tableCollection.FindOne(ctxWithTimeout, bson.M{"table_id": order.Table_id}).Decode(&table)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "message: Table was not found"})
			}
		}

		order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		order.ID = bson.NewObjectID()
		order.Order_id = order.ID.Hex()

		result, err := orderCollection.InsertOne(ctxWithTimeout, order)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "order item was not created"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var order models.Order
		var table models.Table

		var updateObj bson.D

		orderID := ctx.Param("order_id")
		if err := ctx.BindJSON(&order); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if order.Table_id != nil {
			err := menuCollection.FindOne(ctxWithTimeout, bson.M{"table_id": order.Table_id}).Decode(&table)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "message: table was not found"})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "menu", Value: order.Table_id})
		}

		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: order.Updated_at})

		upsert := true
		filter := bson.M{
			"order_id": orderID,
		}
		opt := options.UpdateOne().SetUpsert(upsert)

		result, err := orderCollection.UpdateOne(
			ctxWithTimeout,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			opt,
		)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "order item update failed"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func OrderItemOrderCreator(order models.Order) (string, error) {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	order.ID = bson.NewObjectID()
	order.Order_id = order.ID.Hex()
	_, err := orderCollection.InsertOne(ctxWithTimeout, order)
	if err != nil {
		return "", err
	}

	return order.Order_id, nil

}
