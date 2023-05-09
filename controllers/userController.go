package controllers

import (
	"context"
	"errors"
	"fmt"
	"go-jwt/database"
	helper "go-jwt/helpers"
	"go-jwt/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var usercollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

var errUserNotCreated = errors.New("error in creating user")

func VerifyPassword(dbPwd *string, providedPwd *string) (bool, string) {
	if err := bcrypt.CompareHashAndPassword([]byte(*providedPwd), []byte(*dbPwd)); err != nil {
		return false, fmt.Sprint("email or password is incorrect")
	} else {
		return true, ""
	}
}

func HashPassword(pwd *string) string {
	var hashedBytes []byte
	var err error
	if hashedBytes, err = bcrypt.GenerateFromPassword([]byte(*pwd), 14); err != nil {
		log.Fatal("Error", err)
	}
	return string(hashedBytes)
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		err := usercollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("signup getting called")
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		user := models.User{}
		defer cancel()
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// need to validate the user data which we have got in the request body
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		count, err := usercollection.CountDocuments(ctx, bson.M{"email": user.Email, "$or": bson.A{bson.M{"phone": user.Phone}}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Panic(err)
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "given email/Phone already exist"})
		}
		log.Println("before password hash")
		password := HashPassword(user.Password)
		log.Println("after password hash")
		user.Password = &password
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		helper.GenerateAllTokens(&user)
		log.Println("after generate tokens")
		resultInsertionNumber, err := usercollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errUserNotCreated})
		}
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		user := models.User{}
		foundUser := models.User{}
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		if err := usercollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		if isPasswordValid, msg := VerifyPassword(user.Password, foundUser.Password); !isPasswordValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		helper.GenerateAllTokens(&foundUser)
		helper.UpdateAllTokens(foundUser.Token, foundUser.Refresh_token, foundUser.User_id)
		if err := usercollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}
		projectStage := bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "total_count", Value: 1},
			{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}}}}}

		result, err := usercollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while fetchings users list"})
		}
		var allUsers []bson.M
		// var allUsers []interface{}
		if err := result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allUsers[0])
	}
}
