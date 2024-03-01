package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func Routes(engine *gin.Engine, dbConn *pgx.Conn) {
	// dbConfig := &database.Config{
	// 	DB:      dbConn,
	// 	Queries: database.New(dbConn),
	// }

	// Default route for health check
	engine.GET("/health-check/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Up and Running!",
		})
	})

	// Public API Routes
	// pubRouter := engine.Group("/api/v1/")
	// authRouter := pubRouter.Group("")
	// authRouter.Use(JWTAuth((dbConfig)))

	// Non auth routes
	// pubRouter.POST("register/", SignUp(dbConfig))
	// pubRouter.POST("login/", Login(dbConfig))
	// pubRouter.POST("refresh-token/", RefreshAccessToken(dbConfig))

	// Auth routes
	// authRouter.GET("profile/", GetUserProfile(dbConfig))
	// authRouter.PATCH("profile/", UpdateUserProfile(dbConfig))
	// authRouter.DELETE("profile/", DeleteUserProfile(dbConfig))
	// authRouter.PUT("change-password/", ChangePassword(dbConfig))
}
