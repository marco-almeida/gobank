package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.String()
		c.Next()
		end := time.Now()
		latency := end.Sub(start)

		log.Info().
			Int("status", c.Writer.Status()).
			Str("duration", latency.String()).
			Str("ip", c.ClientIP()).
			Str("method", c.Request.Method).
			Str("path", path).
			Send()
	}
}
