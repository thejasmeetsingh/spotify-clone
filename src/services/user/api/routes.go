package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/thejasmeetsingh/spotify-clone/src/services/user/database"
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

	// Public API Routes
	pubRouter := engine.Group("/api/v1/")
	authRouter := pubRouter.Group("")
	authRouter.Use(JWTAuth((dbConfig)))

	// Non auth routes
	pubRouter.POST("register/", signUp(dbConfig))
	pubRouter.POST("login/", login(dbConfig))
	pubRouter.POST("refresh-token/", refreshAccessToken)

	// Auth routes
	authRouter.GET("profile/", getUserProfile)
	authRouter.PATCH("profile/", updateUserProfile(dbConfig))
	authRouter.DELETE("profile/", deleteUserProfile(dbConfig))
	authRouter.PUT("change-password/", changePassword(dbConfig))
}
