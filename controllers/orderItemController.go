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

type OrderItemPack struct {
	Table_id    *string
	order_items []models.OrderItem
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := orderItemCollection.Find(context.Background(), bson.M{})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing ordered items"})
			return
		}

		var allOrderItems []bson.M
		if err = result.All(ctxWithTimeout, &allOrderItems); err != nil {
			log.Fatal(err)
			return
		}

		ctx.JSON(http.StatusOK, allOrderItems)
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderID := ctx.Param("order_id")
		allOrderItems, err := ItemsByOrder(orderID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items by order ID"})
			return
		}

		ctx.JSON(http.StatusOK, allOrderItems)
	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		orderItemID := ctx.Param("order_item_id")
		var orderItem models.OrderItem

		err := orderItemCollection.FindOne(ctxWithTimeout, bson.M{"order_item_id": orderItemID}).Decode(&orderItem)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order item"})
			return
		}
		ctx.JSON(http.StatusOK, orderItem)
	}
}

func CreateOrderItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var orderItemPack OrderItemPack
		var order models.Order

		if err := ctx.BindJSON(&orderItemPack); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order.Order_Date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		orderItemsToBeInserted := []interface{}{}
		order.Table_id = orderItemPack.Table_id
		order_id, err := OrderItemOrderCreator(order)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for _, orderItem := range orderItemPack.order_items {
			orderItem.Order_id = order_id
			validationErr := validate.Struct(orderItem)
			if validationErr != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}

			orderItem.ID = bson.NewObjectID()
			orderItem.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Order_item_id = orderItem.ID.Hex()
			var num = toFixed(*orderItem.Unit_price, 2)
			orderItem.Unit_price = &num
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
		}

		insertedOrderItems, err := orderItemCollection.InsertMany(ctxWithTimeout, orderItemsToBeInserted)
		if err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, insertedOrderItems)
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var orderItem models.OrderItem
		orderItemID := ctx.Param("order_item_id")
		filter := bson.M{"order_item_id": orderItemID}

		var updateObj bson.D

		if orderItem.Unit_price != nil {
			updateObj = append(updateObj, bson.E{
				Key:   "unit_price",
				Value: orderItem.Unit_price,
			})
		}

		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{
				Key:   "quantity",
				Value: orderItem.Quantity,
			})
		}

		if orderItem.Food_id != nil {
			updateObj = append(updateObj, bson.E{
				Key:   "food_id",
				Value: *orderItem.Food_id,
			})
		}
		orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{
			Key:   "updated_at",
			Value: orderItem.Updated_at,
		})

		upsert := true
		opt := options.UpdateOne().SetUpsert(upsert)

		result, err := orderItemCollection.UpdateOne(
			ctxWithTimeout,
			filter,
			bson.D{
				{
					Key:   "&set",
					Value: updateObj,
				},
			},
			opt,
		)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Order item update failed"})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func ItemsByOrder(id string) (orderItems []bson.M, err error) {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	matchStage := bson.D{
		{Key: "$match", Value: bson.D{{Key: "order_id", Value: id}}},
	}

	lookupFoodStage := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "food"},
			{Key: "localField", Value: "food_id"},   // orderItem food_id argument
			{Key: "foreignField", Value: "food_id"}, // food food_id argument
			{Key: "as", Value: "food"},
		}},
	}

	unwindFoodStage := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$food"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}

	lookupOrderStage := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "order"},
			{Key: "localField", Value: "order_id"},
			{Key: "foreignField", Value: "order_id"},
			{Key: "as", Value: "order"},
		}},
	}

	unwindOrderStage := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$order"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}

	lookupTableStage := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "order"},
			{Key: "localField", Value: "order.table_id"},
			{Key: "foreignField", Value: "table_id"},
			{Key: "as", Value: "table"},
		}},
	}

	unwindTableStage := bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$table"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	}

	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "amount", Value: "$food.price"},
			{Key: "total_count", Value: 1},
			{Key: "food_name", Value: "$food.name"},
			{Key: "food_image", Value: "$food.food_image"},
			{Key: "table_number", Value: "$table.table_number"},
			{Key: "table_id", Value: "$table.table_id"},
			{Key: "order_id", Value: "$order.order_id"},
			{Key: "price", Value: "$food.price"},
			{Key: "quantity", Value: 1},
		}},
	}

	groupStage := bson.D{{
		Key: "$group",
		Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "order_id", Value: "$order_id"},
				{Key: "table_id", Value: "$table_id"},
				{Key: "table_number", Value: "$table_number"},
			}},
			{Key: "payment_due", Value: bson.D{
				{Key: "$sum", Value: "$amount"},
			}},
			{Key: "total_count", Value: bson.D{
				{Key: "$sum", Value: 1},
			}},
			{Key: "order_items", Value: bson.D{
				{Key: "$push", Value: "$$ROOT"},
			}},
		},
	}}

	projectStage2 := bson.D{{
		Key: "$project",
		Value: bson.D{
			{Key: "id", Value: 1},
			{Key: "payment_due", Value: 1},
			{Key: "total_count", Value: 1},
			{Key: "table_number", Value: "$_id.table_number"},
			{Key: "order_items", Value: 1},
		},
	}}

	result, err := orderItemCollection.Aggregate(
		ctxWithTimeout,
		mongo.Pipeline{
			matchStage,
			lookupFoodStage,
			unwindFoodStage,
			lookupOrderStage,
			unwindOrderStage,
			lookupTableStage,
			unwindTableStage,
			projectStage,
			groupStage,
			projectStage2,
		},
	)
	if err != nil {
		panic(err)
	}

	if err = result.All(ctxWithTimeout, &orderItems); err != nil {
		panic(err)
	}

	return orderItems, err
}
