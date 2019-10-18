package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	//SQLConnection is a stablished connection with the netrp relational database
	SQLConnection *storage.SQL
)

//Route defines all the endpoints of this API.
func Route() *gin.Engine {
	r := mlhandlers.DefaultMeliRouter()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	ct := controllers.NewConfigurationController(SQLConnection)

	//POST to /configurations performs a release process configuration create
	r.POST("/configurations", func(c *gin.Context) {
		ct.Create(c)
	})

	//GET to /configurations/:repoName performs a release process configuration get
	r.GET("/configurations/:repoName", func(c *gin.Context) {
		ct.Show(c)
	})

	//GET to /inserts to get the inserts list to populate the new db
	r.GET("/inserts", func(c *gin.Context) {
		ct.GetDBInserts(c)
	})

	//GET to /performance to get the performance values for an application
	r.POST("/performance", func(c *gin.Context) {
		ct.GetAppPerformance(c)
	})

	//PUT to /configurations/:repoName performs a release process configuration update
	r.PUT("/configurations/:repoName", func(c *gin.Context) {
		ct.Update(c)
	})

	//DELETE to /configurations/:repoName performs a release process configuration delete
	r.DELETE("/configurations/:repoName", func(c *gin.Context) {
		ct.Delete(c)
	})

	return r
}
