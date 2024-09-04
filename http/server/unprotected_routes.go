package server

import (
	"phcmis/http/handler"
	m "phcmis/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

func addUnprotectedRoutes(r *gin.Engine, h *handler.Handler, lmt *redis_rate.Limiter) *gin.Engine {
	v1 := r.Group("v1/")

	v1.POST("activate", m.RateLimit(lmt, redis_rate.PerHour(3)), h.ActivateUserAccount)

	v1.POST("login", m.RateLimit(lmt, redis_rate.PerMinute(3)), h.UserLogin)
	v1.POST("user/register", h.CreateUserAccount)
	v1.GET("refresh", h.RenewAcessToken)
	v1.POST("password/forgot", m.RateLimit(lmt, redis_rate.PerHour(4)), h.ForgotPassword)
	v1.POST("password/reset", m.RateLimit(lmt, redis_rate.PerHour(3)), h.ResetPassword)
	v1.GET("health", h.Status)
	return r
}
