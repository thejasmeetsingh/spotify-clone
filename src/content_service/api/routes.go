package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/thejasmeetsingh/spotify-clone/src/content_service/database"
)

func Routes(engine *gin.Engine, dbConn *pgx.Conn) {
	dbConfig := &database.Config{
		DB:      dbConn,
		Queries: database.New(dbConn),
	}

	// Default route for health check
	engine.GET("/health-check/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Up and Running!",
		})
	})

	router := engine.Group("/api/v1/")
	router.Use(JWTAuth((dbConfig)))
}
