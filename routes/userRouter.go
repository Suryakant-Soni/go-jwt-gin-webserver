package routes

import (
	controller "go-jwt/controllers"
	"go-jwt/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	// router.Use(middleware.Authenticate())
	authorizedRoutes := router.Group("/users", middleware.Authenticate())
	authorizedRoutes.GET("", controller.GetUsers())
	authorizedRoutes.GET("/:user_id", controller.GetUser())
}
