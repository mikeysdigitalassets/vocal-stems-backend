package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type IsolateVocalsRequest struct {
	URL string `json:"url"`
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

	r.POST("/api/isolate-vocals", func(c *gin.Context) {
		var request IsolateVocalsRequest

		// binds incoming json url to IsolateVocalsRequest struct
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		log.Println("Processing URL:", request.URL)

		downloadsDir := "downloads"
		outputDir := "output"

		// make sure directories exsists, although they wont since i clean them up
		if err := os.MkdirAll(downloadsDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create downloads directory", "details": err.Error()})
			return
		}
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create output directory", "details": err.Error()})
			return
		}

		// 1 download and convert audio using yt-dlp
		log.Println("Running yt-dlp to download the audio...")
		ytDlpCmd := exec.Command("C:\\Users\\mcros\\myenv\\Scripts\\yt-dlp.exe", request.URL, "-x", "--audio-format", "mp3", "--ffmpeg-location", "C:\\Users\\mcros\\Desktop\\webdev\\ffmpeg-master-latest-win64-gpl\\bin", "-o", filepath.Join(downloadsDir, "%(title)s.%(ext)s"))

		output, err := ytDlpCmd.CombinedOutput()
		log.Println("yt-dlp output:", string(output))
		if err != nil {
			log.Println("yt-dlp error:", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to download and convert video",
				"details": err.Error(),
				"output":  string(output),
			})
			return
		}

		// find the downloaded file
		files, err := filepath.Glob(filepath.Join(downloadsDir, "*.mp3"))
		if err != nil || len(files) == 0 {
			log.Println("Failed to find downloaded file.")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find downloaded audio file"})
			return
		}
		downloadedFile := files[0]

		log.Println("Downloaded file:", downloadedFile)

		// 2 run spleeter to isolate vocals
		log.Println("Running Spleeter to isolate vocals...")
		spleeterCmd := exec.Command("C:\\Users\\mcros\\myenv\\Scripts\\spleeter.exe", "separate", "-o", outputDir, downloadedFile)

		output, err = spleeterCmd.CombinedOutput()
		log.Println("Spleeter output:", string(output))
		if err != nil {
			log.Println("Spleeter error:", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to isolate vocals",
				"details": err.Error(),
				"output":  string(output),
			})
			return
		}

		// 3 find the isolated vocals file
		baseFilename := strings.TrimSuffix(filepath.Base(downloadedFile), ".mp3")
		vocalsFile := filepath.Join(outputDir, baseFilename, "vocals.wav")
		instrumentalsFile := filepath.Join(outputDir, baseFilename, "accompaniment.wav")

		if _, err := os.Stat(vocalsFile); os.IsNotExist(err) {
			log.Println("Failed to find isolated vocals file.")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find isolated vocals file"})
			return
		}

		// 4 stream the file for download
		downloadType := c.Query("type") // 'vocals' or 'instrumentals'
		if downloadType == "instrumentals" {
			c.Writer.Header().Set("Content-Disposition", "attachment; filename="+baseFilename+"_instrumentals.wav")
			c.Writer.Header().Set("Content-Type", "audio/wav")
			c.File(instrumentalsFile)
		} else {
			// default to vocals
			c.Writer.Header().Set("Content-Disposition", "attachment; filename="+baseFilename+"_vocals.wav")
			c.Writer.Header().Set("Content-Type", "audio/wav")
			c.File(vocalsFile)
		}

		go func() {
			log.Println("Cleaning up files...")

			// deletes local files from the the download
			err = os.RemoveAll(downloadsDir)
			if err != nil {
				log.Println("Failed to delete downloads directory:", err.Error())
			}

			err = os.RemoveAll(outputDir)
			if err != nil {
				log.Println("Failed to delete output directory:", err.Error())
			}

			log.Println("Cleanup completed.")
		}()
	})

	r.Run(":5000")
}
