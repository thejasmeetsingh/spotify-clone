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

	pubRouter := engine.Group("/api/v1/")
	authRouter := pubRouter.Group("")
	authRouter.Use(JWTAuth((dbConfig)))

	// Non auth routes
	pubRouter.GET("list/", getContentList(dbConfig))
	pubRouter.GET(":id/", getContentDetail(dbConfig))

	// Auth routes
	authRouter.GET("user/", getUserContentList(dbConfig))
	authRouter.POST("add/", addContent(dbConfig))
	authRouter.PATCH(":id/", updateContent(dbConfig))
	authRouter.PUT(":id/", updateContentS3Key(dbConfig))
	authRouter.DELETE(":id/", deleteContent(dbConfig))
	authRouter.POST("upload-url/", getPresignedURL)
}
