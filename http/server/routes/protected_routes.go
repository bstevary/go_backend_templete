package routes

import (
	"github.com/bstevary/hexagonal/http/handler"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

func addProtectedRoutes(r *gin.Engine, h *handler.Handler, lmt *redis_rate.Limiter) *gin.Engine {

	r.POST("password/change", h.ChangePassword)
	r.GET("auth/logout", h.Logout)

	// user Account

	r.GET("user", h.ListUsers)

	return r
}
