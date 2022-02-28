package http

import (
	"fmt"
	"ryzenlo/to2cloud/internal/models"
	"ryzenlo/to2cloud/internal/pkg/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

func isLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := extraTokenFromRequest(c)
		_, err := auth.VerifyToken(t)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}
		c.Next()
	}
}

func isRootUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := extraTokenFromRequest(c)
		userInfo, err := auth.GetDataFromToken(t)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}
		userID, ok := userInfo["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "failed to get user id"})
			c.Abort()
			return
		}
		user := models.GetUserByUserID(int(userID))
		if user.ID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "user does not exist"})
			c.Abort()
			return
		}
		if user.IsRoot == 0 {
			c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "user does not have the right"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func extraTokenFromRequest(c *gin.Context) string {
	return c.GetHeader("JWT-Token")
}
