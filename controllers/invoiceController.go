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

type InvoiceViewFormat struct {
	Invoice_id       string
	Payment_method   string
	Order_id         string
	Payment_status   *string
	Payment_due      interface{}
	Table_number     interface{}
	Payment_due_date time.Time
	Order_details    interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := invoiceCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing invoice items"})
		}

		var allInvoices []bson.M
		if err = result.All(ctxWithTimeout, &allInvoices); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		invoiceID := ctx.Param("invoice_id")

		var invoice models.Invoice

		errSingleResult := invoiceCollection.FindOne(ctxWithTimeout, bson.M{"invoice_id": invoiceID})
		if errSingleResult != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing invoice item"})
			return
		}

		var invoiceView InvoiceViewFormat

		allOrderItems, err := ItemsByOrder(invoice.Order_id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while sorting invoice item"})
			return
		}
		invoiceView.Order_id = invoice.Order_id
		invoiceView.Payment_due_date = invoice.Payment_due_date

		invoiceView.Payment_method = "null"
		if invoice.Payment_method != nil {
			invoiceView.Payment_method = *invoice.Payment_method
		}

		invoiceView.Invoice_id = invoice.Invoice_id
		invoiceView.Payment_status = invoice.Payment_status
		invoiceView.Payment_due = allOrderItems[0]["payment_due"]
		invoiceView.Table_number = allOrderItems[0]["table_number"]
		invoiceView.Order_details = allOrderItems[0]["order_items"]

		ctx.JSON(http.StatusOK, invoiceView)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var invoice models.Invoice
		if err := ctx.BindJSON(&invoice); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var order models.Order

		err := orderCollection.FindOne(ctxWithTimeout, bson.M{"order_id": invoice.Order_id}).Decode(&order)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error message: order was not found"})
			return
		}

		status := "PENDING"
		if invoice.Payment_status == nil {
			invoice.Payment_status = &status
		}

		invoice.Payment_due_date, _ = time.Parse(time.RFC3339, time.Now().AddDate(0, 0, 1).Format(time.RFC3339))
		invoice.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		invoice.ID = bson.NewObjectID()
		invoice.Invoice_id = invoice.ID.Hex()
		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		result, insertErr := invoiceCollection.InsertOne(ctxWithTimeout, invoice)
		if insertErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invoice was not created"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var invoice models.Invoice
		var invoiceID = ctx.Param("invoice_id")

		if err := ctx.BindJSON(&invoice); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		filter := bson.M{"invoice_id": invoiceID}
		var updateObj bson.D

		if invoice.Payment_method != nil {
			updateObj = append(updateObj, bson.E{
				Key:   "payment_method",
				Value: invoice.Payment_method},
			)
		}

		if invoice.Payment_status != nil {
			updateObj = append(updateObj, bson.E{
				Key:   "payment_status",
				Value: invoice.Payment_status,
			})
		}

		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{
			Key:   "updated_at",
			Value: invoice.Updated_at,
		})

		upsert := true
		opt := options.UpdateOne().SetUpsert(upsert)

		status := "PENDING"
		if invoice.Payment_status == nil {
			invoice.Payment_status = &status
		}

		result, err := invoiceCollection.UpdateOne(
			ctxWithTimeout,
			filter,
			bson.D{
				{Key: "&set", Value: updateObj},
			},
			opt,
		)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invoice item update failed"})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
