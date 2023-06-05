package routes

import (
	controller "github.com/Pawan109/golang-jwt-project/controllers"

	"github.com/Pawan109/golang-jwt-project/middleware"

	"github.com/gin-gonic/gin"
)

//here we will be required to use a middleware , as after logging in the user has the token->which needs to be protected

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.Getusers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())
}
