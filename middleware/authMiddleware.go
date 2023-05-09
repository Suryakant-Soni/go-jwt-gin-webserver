package middleware

import (
	helper "go-jwt/helpers"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("Token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token not available"})
			c.Abort()
			return
		}
		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			log.Fatal("Error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}
		c.Set("email", claims.Email)
		c.Set("first_name", claims.First_Name)
		c.Set("last_name", claims.Last_Name)
		c.Set("uid", claims.Id)
		c.Set("user_type", claims.User_type)
		c.Next()
	}
}
