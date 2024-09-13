package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/koinav/ecommerce/controllers"
)

func UserRoutes(incoming *gin.Engine) {
	incoming.POST("/users/signup", controllers.SignUp())
	incoming.POST("/users/login", controllers.Login())
	incoming.POST("/admin/addproduct", controllers.ProductViewerAdmin())
	incoming.GET("/users/productview", controllers.SearchProduct())
	incoming.GET("/users/search", controllers.SearchProductByQuery())
}
