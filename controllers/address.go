package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/koinav/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("userID")
		if userID == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid userID"})
			c.Abort()
			return
		}

		id, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Internal error")
			return
		}

		var address models.Address
		address.AddressID = primitive.NewObjectID()

		if err = c.BindJSON(&address); err != nil {
			c.JSON(http.StatusNotAcceptable, err.Error())
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		matchFilter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: id}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$address"}}}}
		group := bson.D{
			{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$address_id"},
				{Key: "count", Value: bson.D{primitive.E{Key: "&sum", Value: 1}}}}},
		}

		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{matchFilter, unwind, group})
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Internal error")
			return
		}

		var addressInfo []bson.M
		if err = pointCursor.All(ctx, &addressInfo); err != nil {
			c.JSON(http.StatusInternalServerError, "Internal error")
			return
		}

		var size int32
		for _, addressNo := range addressInfo {
			count := addressNo["count"]
			size = count.(int32)
		}
		if size < 2 {
			filter := bson.D{primitive.E{Key: "_id", Value: address}}
			update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "address", Value: address}}}}
			_, err := UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, "Not allowed")
			return
		}

		ctx.Done()
		c.JSON(http.StatusOK, "Updated successfully")
	}
}

func EditHomeAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")
		if userID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid"})
			c.Abort()
			return
		}

		id, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "internal error")
			return
		}

		var editAddress models.Address
		if err = c.BindJSON(&editAddress); err != nil {
			c.JSON(http.StatusNotAcceptable, err.Error())
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: id}}
		update := bson.D{{Key: "$set", Value: bson.D{
			primitive.E{Key: "address.0.house_name", Value: editAddress.House},
			primitive.E{Key: "address.0.street_name", Value: editAddress.Street},
			primitive.E{Key: "address.0.city_name", Value: editAddress.City},
			primitive.E{Key: "address.0.post_code", Value: editAddress.PostCode}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Something went wrong")
			return
		}

		ctx.Done()
		c.JSON(http.StatusOK, "Updated successfully")
	}
}

func EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")
		if userID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid"})
			c.Abort()
			return
		}

		id, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "internal error")
			return
		}

		var editAddress models.Address
		if err = c.BindJSON(&editAddress); err != nil {
			c.JSON(http.StatusNotAcceptable, err.Error())
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: id}}
		update := bson.D{{Key: "$set", Value: bson.D{
			primitive.E{Key: "address.1.house_name", Value: editAddress.House},
			primitive.E{Key: "address.1.street_name", Value: editAddress.Street},
			primitive.E{Key: "address.1.city_name", Value: editAddress.City},
			primitive.E{Key: "address.1.post_code", Value: editAddress.PostCode}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Something went wrong")
			return
		}

		ctx.Done()
		c.JSON(http.StatusOK, "Updated successfully")
	}
}

func DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")

		if userID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid search index"})
			c.Abort()
			return
		}

		addresses := make([]models.Address, 0)
		id, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "internal error")
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: id}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotFound, "wrong command")
			return
		}

		ctx.Done()
		c.JSON(http.StatusOK, "deleted successfully")
	}
}
