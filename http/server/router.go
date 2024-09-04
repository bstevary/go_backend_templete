package server

import (
	"phcmis/config"
	"phcmis/databases/redis/daemon"
	_ "phcmis/docs"
	"phcmis/services/auth"

	"github.com/go-redis/cache/v9"
	"github.com/go-redis/redis_rate/v10"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"phcmis/databases/persist/db"

	"phcmis/http/handler"
	"phcmis/http/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	config          config.Config
	db              db.Store
	token           auth.TokenGenerator
	limiter         *redis_rate.Limiter
	taskDistributer daemon.TaskDistributor
	cache           *cache.Cache
}

// newRouter creates a new instance of the gin.Engine router with the provided RouterConfig.
// It sets up middleware, session management, CSRF protection, rate limiting, and routes.
// The router is returned as the result.
func newRouter(rc RouterConfig) *gin.Engine {
	router := gin.New()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173"}
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")

	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.Use(middleware.LoggerMiddleware())
	router.Use(gin.Recovery())

	Handler := handler.NewHandler(rc.db, rc.taskDistributer, rc.cache, rc.token, rc.config)
	router = addUnprotectedRoutes(router, Handler, rc.limiter)
	router.Use(middleware.AuthMiddleware(rc.token))

	router = addProtectedRoutes(router, Handler, rc.limiter)

	return router
}
