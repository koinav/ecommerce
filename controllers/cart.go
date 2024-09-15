package controllers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/koinav/ecommerce/database"
	"github.com/koinav/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"time"
)

type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewApp(prodCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("productID")
		if productQueryID == "" {
			c.JSON(http.StatusBadRequest, errors.New("productID is not set"))
			c.Abort()
			return
		}

		userQueryID := c.Query("userID")
		if userQueryID == "" {
			c.JSON(http.StatusBadRequest, errors.New("userID is not set"))
			c.Abort()
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.AddProductToCart(ctx, app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, "Successfully added to cart")
	}
}

func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("productID")
		if productQueryID == "" {
			c.JSON(http.StatusBadRequest, errors.New("productID is not set"))
			c.Abort()
			return
		}

		userQueryID := c.Query("userID")
		if userQueryID == "" {
			c.JSON(http.StatusBadRequest, errors.New("userID is not set"))
			c.Abort()
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.RemoveCartItem(ctx, app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, "Item removed Successfully")

	}
}

func GetUserCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("userID")
		if userID == "" {
			c.JSON(http.StatusBadRequest, errors.New("userID is not set"))
			c.Abort()
			return
		}

		id, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var cart models.User
		err = UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&cart)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "not found")
			return
		}

		filterMatch := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: id}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$user_cart"}}}}
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$user_cart.price"}}}}}}
		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filterMatch, unwind, grouping})
		if err != nil {
			log.Println(err)
		}

		var listing []bson.M
		if err := pointCursor.All(ctx, &listing); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		for _, json := range listing {
			c.JSON(http.StatusOK, json["total"])
			c.JSON(http.StatusOK, cart.UserCart)
		}
	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.Query("userID")
		if userQueryID == "" {
			c.JSON(http.StatusBadRequest, errors.New("userID is not set"))
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		err := database.BuyItemFromCart(ctx, app.userCollection, userQueryID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		c.JSON(http.StatusOK, "Order placed successfully")
	}
}

func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryID := c.Query("productID")
		if productQueryID == "" {
			c.JSON(http.StatusBadRequest, errors.New("productID is not set"))
			c.Abort()
			return
		}

		userQueryID := c.Query("userID")
		if userQueryID == "" {
			c.JSON(http.StatusBadRequest, errors.New("userID is not set"))
			c.Abort()
			return
		}

		productID, err := primitive.ObjectIDFromHex(productQueryID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.InstantBuy(ctx, app.prodCollection, app.userCollection, productID, userQueryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		}
		c.JSON(http.StatusOK, "Order placed successfully")
	}
}
