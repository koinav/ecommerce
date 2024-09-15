package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/koinav/ecommerce/database"
	"github.com/koinav/ecommerce/models"
	"github.com/koinav/ecommerce/tokens"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

var (
	UserCollection    *mongo.Collection = database.UserData(database.Client, "Users")
	ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
	Validate                            = validator.New()
)

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, givenPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	valid := true
	msg := ""

	if err != nil {
		msg = "login or password is incorrect"
		valid = false
	}

	return valid, msg
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := Validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		}

		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt = user.CreatedAt
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()
		token, refreshToken, err := tokens.TokenGenerator(*user.Email, *user.FirstName, *user.LastName, user.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		user.Token = &token
		user.RefreshToken = &refreshToken
		user.UserCart = make([]models.ProductInCart, 0)
		user.AddressDetails = make([]models.Address, 0)
		user.OrderStatus = make([]models.Order, 0)

		_, err = UserCollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user was not created"})
		}

		c.JSON(http.StatusCreated, "Successfully signed up!")
	}
}

func LogIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user, foundUser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
			return
		}

		isValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !isValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "msg"})
			fmt.Println(msg)
			return
		}

		token, refreshToken, err := tokens.TokenGenerator(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, foundUser.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		err = tokens.UpdateAllTokens(token, refreshToken, foundUser.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		c.JSON(http.StatusFound, foundUser)
	}
}

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var products models.Product

		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		products.ProductID = primitive.NewObjectID()
		_, err := ProductCollection.InsertOne(ctx, products)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not inserted"})
			return
		}

		c.JSON(http.StatusOK, "successfully added")
	}
}

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var productList []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := ProductCollection.Find(ctx, bson.D{})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Something went wrong")
			return
		}

		err = cursor.All(ctx, &productList)

		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		defer cursor.Close(ctx)

		if err := cursor.Err(); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, "invalid")
			return
		}

		c.JSON(http.StatusOK, &productList)
	}
}

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchProducts []models.Product

		queryParam := c.Query("name")
		if queryParam == "" {
			log.Println("query is empty")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search index"})
			c.Abort()
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		searchQuery, err := ProductCollection.Find(ctx, bson.M{"product_name": bson.M{"$regex": queryParam}})
		if err != nil {
			c.JSON(http.StatusNotFound, "something went wrong")
			return
		}

		err = searchQuery.All(ctx, &searchProducts)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, "invalid")
			return
		}
		defer searchQuery.Close(ctx)

		if err := searchQuery.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "invalid request")
			return
		}

		c.IndentedJSON(http.StatusOK, searchProducts)
	}
}
