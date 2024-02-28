package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
	"github.com/thejasmeetsingh/spotify-clone/src/user_service/api"
)

func getLoggerFormat(params gin.LogFormatterParams) string {
	return fmt.Sprintf(
		"%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
		params.ClientIP,
		params.TimeStamp.Format(time.RFC1123),
		params.Method,
		params.Path,
		params.Request.Proto,
		params.StatusCode,
		params.Latency,
		params.Request.UserAgent(),
		params.ErrorMessage,
	)
}

func main() {
	engine := gin.Default()

	// DB config
	conn, err := pgx.Connect(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalln("error while connecting to DB: ", err)
	}
	defer conn.Close(context.Background())

	// Load API routes
	api.Routes(engine, conn)

	// Server config
	engine.Use(gin.LoggerWithFormatter(getLoggerFormat))

	s := &http.Server{
		Addr:           ":" + os.Getenv("HTTP_PORT"),
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Infoln("HTTP service is up & running")

	if err := s.ListenAndServe(); err != nil {
		log.Fatalln("failed to serve HTTP: ", err)
	}
}
