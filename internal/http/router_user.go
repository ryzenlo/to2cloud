package http

import (
	"fmt"
	"net/http"
	"ryzenlo/to2cloud/internal/models"
	"ryzenlo/to2cloud/internal/pkg/auth"

	"github.com/gin-gonic/gin"
)

type UserLoginParam struct {
	Name string `json:"username" binding:"required"`
	Pwd  string `json:"userpwd" binding:"required"`
}

type UserParam struct {
	Name     string `json:"username" binding:"required"`
	Pwd      string `json:"userpwd" binding:"required"`
	Nickname string `json:"nickname"`
	Status   int    `json:"status"`
}

func AddUser(c *gin.Context) {
	var userParam UserParam
	if err := c.ShouldBindJSON(&userParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	dbUser := models.GetUserByName(userParam.Name)
	if dbUser.ID > 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "user name exists"})
		return
	}
	user := &models.User{
		Username: userParam.Name,
		Password: userParam.Pwd,
		Nickname: userParam.Nickname,
	}
	if err := models.AddUser(user); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("failed to create user,%v", err)})
		return
	}
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func userLogin(c *gin.Context) {
	var loginParam UserLoginParam
	if err := c.ShouldBindJSON(&loginParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	theUser := models.GetUserByName(loginParam.Name)
	if theUser.ID == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "invalid user name or password"})
		return
	}
	if !models.CheckPassword(theUser.Password, loginParam.Pwd) {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "invalid user name or password"})
		return
	}
	token, err := auth.CreateToken(int64(theUser.ID), theUser.IsRoot)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "", "data": gin.H{"token": token}})
}

func GetUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user := models.GetUserByUserID(userID.(int))
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "", "data": user})
}
