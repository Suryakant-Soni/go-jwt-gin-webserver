package routes

import (
	controller "go-jwt/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("/users/signup", controller.Signup())
	router.POST("/users/login", controller.Login())
}
