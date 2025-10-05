package server

import (
	"net/http"
	"time"

	"github.com/bstevary/hexagonal/config"
	"github.com/bstevary/hexagonal/utils/auth"

	"github.com/go-redis/redis_rate/v10"

	"github.com/bstevary/hexagonal/http/handler"
	"github.com/bstevary/hexagonal/http/middleware"
	"github.com/bstevary/hexagonal/http/server/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type routerConfig struct {
	handler *handler.Handler
	env     *config.Env
	token   auth.TokenGenerator
	limiter *redis_rate.Limiter
}

func newRouter(rc routerConfig) *gin.Engine {
	router := gin.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     rc.env.AllowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Accept", "Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware())

	v1 := router.Group("api/v1/")

	routes.AddAuthRoutes(v1, rc.handler, rc.limiter)

	// authenticated
	v1.Use(middleware.AuthMiddleware(rc.token))

	return router
}
