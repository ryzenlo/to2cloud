package http

import "github.com/gin-gonic/gin"

func GetHttpRouter() *gin.Engine {
      r := gin.Default()
	  setupRoutes(r)
	  return r
}
