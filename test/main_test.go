package test

import (
	"context"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	config "github.com/bstevary/hexagonal/config"
	"github.com/bstevary/hexagonal/database/db"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testStore db.Database

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	config, err := config.LoadEnv("../")
	if err != nil {
		log.Fatal().Msgf("cannot load configuration: %v", err)
	}
	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Msgf("cannot connect to db: %v", err)
	}
	testStore = db.NewDatabase(connPool)
	os.Exit(m.Run())
}
