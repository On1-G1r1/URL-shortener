package domain

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var urlStore = make(map[string]string)

func main() {
	r := gin.Default()

	r.POST("/shorten", func(c *gin.Context) {
		var req struct {
			OriginalURL `json:"original_url" binding:"required"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		ShortURL := GenerateShortURL(req.OriginalURL)
		urlstore[ShortURL] = req.OriginalURL

		c.JSON(http.StatusOK, gin.H{
			"short_code": ShortURL,
			"short_url":  "http://localhost:8080/" + ShortURL,
		})

	})
	r.GET("/:code", func(c *gin.Context) {
		shortCode := c.Param("shortURL")
		originalURL, exist := urlStore[shortCode]
		if !exist {
			c.JSON(http.StatusNotFound, gin.H{"error": "short URL not found"})
			return
		}
		c.Redirect(http.StatusFound, originalURL)
	})
	r.Run(":8080")

}
