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

	router := engine.Group("/api/v1/")

	router.POST("register/", SignUp(dbConfig))
	router.POST("login/", Login(dbConfig))
	router.POST("refresh-token/", RefreshAccessToken(dbConfig))
}
