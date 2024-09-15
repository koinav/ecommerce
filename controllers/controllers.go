package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/koinav/ecommerce/database"
	"github.com/koinav/ecommerce/models"
	"github.com/koinav/ecommerce/tokens"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	UserCollection    = database.UserData(database.Client, "Users")
	ProductCollection = database.ProductData(database.Client, "Products")
	Validate          = validator.New()
)

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func verifyPassword(userPassword string, givenPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	valid := true
	var myErr error

	if err != nil {
		myErr = errors.New("login or password is incorrect")
		valid = false
	}

	return valid, myErr
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
			return
		}

		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
			return
		}

		user.Password = hashPassword(user.Password)

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt = user.CreatedAt
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()
		token, refreshToken, err := tokens.TokenGenerator(user.Email, user.FirstName, user.LastName, user.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		user.Token = token
		user.RefreshToken = refreshToken
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
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
			return
		}
		defer cancel()

		isValid, msg := verifyPassword(user.Password, foundUser.Password)
		if !isValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "msg"})
			fmt.Println(msg)
			return
		}

		token, refreshToken, err := tokens.TokenGenerator(foundUser.Email, foundUser.FirstName, foundUser.LastName, foundUser.UserID)
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
		var product models.Product

		if err := c.BindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		product.ProductID = primitive.NewObjectID()
		_, err := ProductCollection.InsertOne(ctx, product)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not inserted"})
			return
		}

		c.JSON(http.StatusOK, "Successfully added")
	}
}

func ViewProducts() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var productList, err = getAllProducts(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		c.JSON(http.StatusOK, &productList)
	}
}

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		queryParam := c.Query("name")
		if queryParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search index"})
			c.Abort()
			return
		}
		queryParam = strings.ToLower(queryParam)

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var productList, err = getAllProducts(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		var searchResults []models.Product
		for _, product := range productList {
			name := strings.ToLower(product.ProductName)
			if strings.Contains(name, queryParam) {
				searchResults = append(searchResults, product)
			}
		}

		c.JSON(http.StatusOK, &searchResults)
	}
}

func getAllProducts(ctx context.Context) ([]models.Product, error) {
	var productList []models.Product

	cursor, err := ProductCollection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &productList)
	if err != nil {
		return nil, err
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return productList, nil
}
