package helper

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Pawan109/golang-jwt-project/database"
	jwt "github.com/dgrijalva/jwt-go" //just like there's npm/yarn for js-> for GO someone built this driver
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//Your JWT TOKEN uses a hashing mechanism to take the details we give it & gives us a token
//and we can go to the JWT.io website & decode that token & get all details back
//so its just a way to store information
//JWT also consists of a secret key - something thats with you & that's secret
//we will be creating a JWT token using secret key

type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	User_type  string
	jwt.StandardClaims
}

//like before  - again creating userCollection for db here

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY = os.Getenv("SECRET_KEY") //lekin .evn mein bhi secret_key kaha se laayein????

func GenerateAllTokens(email string, firstName string, lastName string, userType string, uid string) (signedToken string, signedRefreshToken string, err error) {

	claims := &SignedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		Uid:        uid,
		User_type:  userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}
	//SigningMethodHS256 ->is an algorithm which generates your token
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}
	return token, refreshToken, err
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {

	token, err := jwt.ParseWithClaims( //this is jwt internal fn -> used to decode token
		//token lega , signedDetails lega -> aur return karega secret key (if we have) .
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) { //function takes the token as argument and returns interfce & error
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}

	//if someone sends us an expired token ->
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired")
		msg = err.Error()
		return
	}

	return claims, msg

}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateObj primitive.D //D is an ordered representation of a BSON document. This type should be used when the order of the elements matter

	updateObj = append(updateObj, bson.E{"token", signedToken}) //E: A single element inside a D type

	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", Updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)

	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}
	return
}
