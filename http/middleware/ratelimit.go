package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"phcmis/services/gin_pgx_err"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

func RateLimit(limiter *redis_rate.Limiter, strategy redis_rate.Limit) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() + c.Request.URL.Path

		res, err := limiter.Allow(c.Request.Context(), key, strategy)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
			return
		}

		// Set standard rate limit headers
		c.Header("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		c.Header("X-RateLimit-Limit", strconv.Itoa(res.Limit.Rate))

		if res.Allowed == 0 {
			// We are rate limited.
			seconds := int(res.RetryAfter / time.Second)
			resetTime := time.Now().Add(res.RetryAfter).Unix() // Calculate reset time in UTC epoch seconds
			c.Header("X-RateLimit-RetryAfter", strconv.Itoa(seconds))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

			// Stop processing and return the error.
			err := fmt.Errorf("rate limit exceeded, retry after %d seconds", seconds)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin_pgx_err.ErrorResponse(err))
			return
		}

		// Continue processing as normal.
		c.Next()
	}
}
