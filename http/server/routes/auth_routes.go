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

	r.POST("login", m.RateLimit(lmt, redis_rate.PerMinute(3)), h.UserLogin)
	r.GET("logout", m.RateLimit(lmt, redis_rate.PerMinute(3)), h.Logout)

	r.POST("mfa", m.RateLimit(lmt, redis_rate.PerMinute(3)), h.MFAChallenge)
	r.GET("refresh", m.RateLimit(lmt, redis_rate.PerMinute(3)), h.RefreshToken)

	r.POST("password/forgot", m.RateLimit(lmt, redis_rate.PerSecond(1)), h.ForgotPassword)
	r.POST("password/reset", m.RateLimit(lmt, redis_rate.PerHour(10)), h.ResetUserAccount)
	r.POST("password/reset/otp", m.RateLimit(lmt, redis_rate.PerHour(10)), h.ResetPassword)

}
