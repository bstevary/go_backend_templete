package routes

import (
	"github.com/bstevary/hexagonal/http/handler"
	m "github.com/bstevary/hexagonal/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

func AddAuthRoutes(r *gin.RouterGroup, h *handler.Handler, lmt *redis_rate.Limiter) {
	r.GET("health", h.Status)

	onboarding := r.Group("onboarding/")
	onboarding.POST("user", m.RateLimit(lmt, redis_rate.PerHour(10)), h.CreateUserAccount)
	onboarding.POST("verify", m.RateLimit(lmt, redis_rate.PerHour(10)), h.ActivateUserAccount)

	auth := r.Group("auth/")
	auth.POST("login", m.RateLimit(lmt, redis_rate.PerMinute(3)), h.UserLogin)
	auth.GET("logout", m.RateLimit(lmt, redis_rate.PerMinute(3)), h.Logout)

	auth.POST("mfa", m.RateLimit(lmt, redis_rate.PerMinute(3)), h.MFAChallenge)
	auth.GET("refresh", m.RateLimit(lmt, redis_rate.PerMinute(3)), h.RenewAccessToken)

	auth.POST("password/forgot", m.RateLimit(lmt, redis_rate.PerSecond(1)), h.ForgotPassword)
	auth.POST("password/reset", m.RateLimit(lmt, redis_rate.PerHour(10)), h.ResetPassword)

}
