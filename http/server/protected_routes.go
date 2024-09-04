package server

import (
	"phcmis/http/handler"
	m "phcmis/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

func addProtectedRoutes(router *gin.Engine, h *handler.Handler, lmt *redis_rate.Limiter) *gin.Engine {
	{

		v1 := router.Group("/v1/", m.RateLimit(lmt, redis_rate.PerSecond(6)))

		{
			v1.POST("password/change", h.ChangePassword)
			v1.GET("logout", h.Logout)

			// user Account
			v1.PATCH("user/:phc_id", h.UpdateUser)
			v1.DELETE("user/:phc_id", h.DeleteUser)
			v1.GET("user/:phc_id", h.GetUser)
			v1.GET("users", h.ListUsers)

		}
		return router
	}
}
