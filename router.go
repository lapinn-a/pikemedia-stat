package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func Logging(c *gin.Context) {
	log.Printf("%v %v", c.Request.Method, c.Request.RequestURI)
	c.Next()
}

func (stat *Stat) Router() *gin.Engine {
	router := gin.New()
	router.Use(Logging)
	router.GET("/ping", stat.Ping)
	router.GET("/stat", stat.Stats)
	router.POST("/collect", stat.Collect)
	router.GET("/report", stat.Report)
	return router
}
