package routes

import (
	controller "github.com/Pawan109/golang-jwt-project/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("users/signup", controller.Signup())
	incomingRoutes.POST("users/login", controller.Login())

}

///authRoutes do not have any middleware here , as during login signup no protection is required as the user doesn't has any kinda token
