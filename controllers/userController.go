package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Pawan109/golang-jwt-project/database"
	helper "github.com/Pawan109/golang-jwt-project/helpers"
	"github.com/Pawan109/golang-jwt-project/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10" //bcz we would need a validator
	"golang.org/x/crypto/bcrypt"             //bcz would need to encrypt our password

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//userCollection ke db ko open karo -> using OpenCollection fn that we wrote in dbconnection
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var validate = validator.New()

//in a db you cannot store users password as is -> you have to hash it
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

//jis password se user ne sign in kiya tha wo , aur jo usne abhi enter kiya wo
//matches- yes/no  msg -> password is incorrrect....blabla
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true //flag
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("email or password is incorrect")
		check = false
	}

	return check, msg
}

//signup function mein -> first we care converting the JSON we wrote in models to struct with bindJSON
//then we are validating if all the required fields are there
//then we are checking if any user is trying to log in with  the phone/email which has already been used ,
//we do this by maintaining count , of each email.phone
//only after signup we will generate and assign token -> which we also do in this function
func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second) //100 sec mein session timesout
		var user models.User                                                         //jaha json mein likha hai

		if err := c.BindJSON(&user); err != nil { //bind json ->deserialises the JSON/XML..into a struct and
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
		}

		password := HashPassword(*user.Password) //user sign in kiya -> usne jo passowrd set kiya=> usko hash krke it sets
		user.Password = &password                //and updates it with the hashed password .

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone no. already exists"})

		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, *&user.User_id)
		user.Token = &token //setting the token value after generating
		user.Refresh_token = &refreshToken

		//now this user has signed up and been assigned a token -> now i'll put it in the db
		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

//login fn mein -user have already signed up , so we will find his account by searching his mail in the db
//it also checks if passwordisValid using verifypassword fn
//now if theres no error it'll find the user from the db using his userid
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)

		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return

		}

		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}

		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, foundUser)

	}
}

//admin wants to get details of the users
//recordsPerPage & min no. of page should be 1
//$match $group $project use krke we will get the alluserslist and then we will put it in a slice
func Getusers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		//ek page mein there should be min 10 recordds by default
		//and there should be minimum one page
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		//the matchStage (the pipeline stage) basically just matches the values you put in the bson.D{{....}}-> and it matches it with "$match"
		//ex: https://www.mongodb.com/docs/manual/core/aggregation-pipeline/ -> see the pizza example
		//matchStage & groupStage often used together - are pipeline functions
		// $sum is not a pipeline function
		//if you want to find out the count of all records in the db -> group them
		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}},        // this groups the db by id  -> which will help us create a total count of all the ids
			{"total_count", bson.D{{"$sum", 1}}},    //$sum helps to calculate the sum of all records -> basically $sum ,1 -> is used to count the total no. of records
			{"data", bson.D{{"$push", "$$ROOT"}}}}}} //if you don't do $push - you aren't going to see the data , you're just going to see the count
		//$project is used to add/remove from the list of data ... or can also do  some queries
		//$slice - returns array elemets -(first three, last three ... like that)
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},         ///mtlb id exclude krke
				{"total_count", 1}, //displaying total count .
				{"usr_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})

		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})

		}
		var allusers []bson.M //M : An unordered representation of a BSON document (map)
		if err = result.All(ctx, &allusers); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allusers[0])
	}
}

//only an admin can get access to the regular db & not a regular user
//so means that if in this api if its the data of another user then a regular user  cannot access it only the admin can
//gin gives us access to the handler fn -> if you don't use gin then you use http.Handlefunc..

//so basically GetUser() takes user_id from postman -> checks if user_id & user_type(admin or regular user) matches or not
//and finally searches for the user_id in the userCollection db and then from there we can get the user
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id") //so with the help of c you can access the parameters sent from the postman

		//i have to call the helper fn to check if the user is admin or not
		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		//now to get the user
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user) //userCollection db se koi user-ID nikal re hai
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
