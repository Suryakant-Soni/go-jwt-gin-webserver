package main

import (
	"fmt"
	routes "go-jwt/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	// gin.SetMode(gin.ReleaseMode)
	if port == "" {
		port = "8000"
	}
	log.Println("main file called")
	router := gin.New()
	router.Use(gin.Logger())
	// we will just add our routes from the separate files into the same gin.engine object
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-get1", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"sucess": "access granted for api-1"})
		fmt.Println("apiget1 called")
	})

	router.GET("/api-get2", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"sucess": "access granted for api-2"})
	})
	log.Printf("port %v", port)
	router.Run(":" + port)
}
