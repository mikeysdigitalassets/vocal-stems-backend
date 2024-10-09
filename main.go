package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// struct for incoming post request body
type IsolateVocalsRequest struct {
	URL string `json: "url"` // url field to hold youtube url
}

func main() {
	r := gin.Default()

	// cors bullshit
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://vocal-stems.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// post to handle vocal isolation request
	r.POST("/api/isolate-vocals", func(c *gin.Context) {
		var request IsolateVocalsRequest

		// bind incoming json youtube url to isolatevocals struct
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// for now this is just dummy download link for deving
		downloadLink := "https://vocal-stems.com/download/isolated-vocals.mp3"

		// respond with json containing downloadlink
		c.JSON(http.StatusOK, gin.H{
			"downloadUrl": downloadLink,
		})

	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to vocal stems backend",
		})
	})
	r.Run(":5000")
}
