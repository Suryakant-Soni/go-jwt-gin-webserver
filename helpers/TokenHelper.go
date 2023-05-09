package helper

import (
	"context"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go-jwt/database"
	"go-jwt/models"
)

type SignedDetails struct {
	Email      string
	First_Name string
	Last_Name  string
	Uid        string
	User_type  string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(user *models.User) {
	claims := &SignedDetails{
		Email:      *user.Email,
		First_Name: *user.First_Name,
		Last_Name:  *user.Last_Name,
		Uid:        user.ID.String(),
		User_type:  *user.User_type,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Fatal("Error in claims token", err)
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Fatal("Error in refresh token", err)
	}
	user.Token = &token
	user.Refresh_token = &refreshToken
}

func UpdateAllTokens(signedToken *string, signedRefreshToken *string, userId string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	var updateObj primitive.D
	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})
	Updated_at, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		log.Fatal("Error in time parsing", err) // it prints and then exits with value 1
	}
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: Updated_at})
	upsert := true
	filer := bson.M{"user_id": userId}
	options := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err = userCollection.UpdateOne(
		ctx,
		filer,
		bson.D{
			{Key: "$set", Value: updateObj},
		},
		&options,
	)
	if err != nil {
		log.Fatal("Error in updating tokens in db", err)
	}
	defer cancel()
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	log.Println("started ValidateToken")
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		msg = "token is invalid"
		msg = err.Error()
		return
	}
	// check if the claims we got back have the same type as of out custom claim type i.e. signedDetails
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "token is invalid"
		msg = err.Error()
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "token is expired"
		msg = err.Error()
		return
	}
	return claims, msg
}
