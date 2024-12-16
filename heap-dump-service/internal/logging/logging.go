package logging

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func SetupLogging() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func GetDurationInMillseconds(start time.Time) float64 {
	end := time.Now()
	duration := end.Sub(start)
	milliseconds := float64(duration) / float64(time.Millisecond)
	rounded := float64(int(milliseconds*100+.5)) / 100
	return rounded
}

func JSONLogMiddleware() gin.HandlerFunc {
	log.SetFormatter(&log.JSONFormatter{})
	return func(c *gin.Context) {
		// Start timer
		//start := time.Now()

		// Process Request
		c.Next()

		// Stop timer
		//duration := GetDurationInMillseconds(start)

		entry := log.WithFields(log.Fields{
			"caller":   c.FullPath(),
			"method":   c.Request.Method,
			"path":     c.Request.RequestURI,
			"status":   c.Writer.Status(),
			"referrer": c.Request.Referer(),
		})

		if c.Writer.Status() >= 500 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("")
		}
	}
}
