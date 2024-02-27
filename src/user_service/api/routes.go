package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/thejasmeetsingh/spotify-clone/src/user_service/database"
)

func Routes(engine *gin.Engine, dbConn *pgx.Conn) {
	dbConfig := &database.Config{
		DB:      dbConn,
		Queries: database.New(dbConn),
	}

	// Public API Routes
	pubRouter := engine.Group("/api/v1/")
	authRouter := pubRouter.Group("")
	authRouter.Use(JWTAuth((dbConfig)))

	// Non auth routes
	pubRouter.POST("register/", SignUp(dbConfig))
	pubRouter.POST("login/", Login(dbConfig))
	pubRouter.POST("refresh-token/", RefreshAccessToken(dbConfig))

	// Auth routes
	authRouter.GET("profile/", GetUserProfile(dbConfig))
	authRouter.PATCH("profile/", UpdateUserProfile(dbConfig))
	authRouter.DELETE("profile/", DeleteUserProfile(dbConfig))
	authRouter.PUT("change-password/", ChangePassword(dbConfig))
}
