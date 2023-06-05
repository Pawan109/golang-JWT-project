package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("user_type") // ya toh admin aaaega ya user
	err = nil
	if userType != role {
		err = errors.New("Unauthorized to access this resource  ")
		return err
	}
	return err
}

func MatchUserTypeToUid(c *gin.Context, userId string) (err error) {
	userType := c.GetString("user_type")
	uid := c.GetString("uid")
	err = nil

	if userType == "USER" && uid != userId { //means woh user ki uid nai hai -> only admin can view others data
		err = errors.New("Unauthorized to access  this resource ")
		return err
	}
	err = CheckUserType(c, userType)
	return err
}
