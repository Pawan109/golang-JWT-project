package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	routes "github.com/Pawan109/golang-jwt-project/routes"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT") //it will assign us a PORT

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	//now we will create 2 APIs - remember when we use gin we don't have to use w & r for req & res
	router.GET("/api-1", func(c *gin.Context) {
		//you can also easily set headers like this -
		c.JSON(200, gin.H{"success": "Access granted for api-1"}) //otherwise pehle w/o gin we had to set headers for response like-> w.Setheader..
		//pros of using such packages is that it saves time
		//but theres a con that you don't really get to learn the underlying basics
	})

	router.GET("/api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for the api-2"})
	})

	router.Run(":" + port)
}
