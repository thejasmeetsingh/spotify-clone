package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/thejasmeetsingh/spotify-clone/src/content_service/database"
	"github.com/thejasmeetsingh/spotify-clone/src/content_service/internal"
)

func JWTAuth(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		headerAuthToken := ctx.GetHeader("Authorization")

		if headerAuthToken == "" {
			ctx.JSON(http.StatusForbidden, gin.H{"message": "Authentication required"})
			ctx.Abort()
			return
		}

		// Split the token string
		authToken := strings.Split(headerAuthToken, " ")

		// Validate the token string
		if len(authToken) != 2 || authToken[0] != "Bearer" {
			ctx.JSON(http.StatusForbidden, gin.H{"message": "Invalid authentication format"})
			ctx.Abort()
			return
		}

		// Fetch user and set to the context
		user, err := internal.GetUserDetail(ctx, authToken[1])
		if err != nil {
			log.Errorln("error while fetching user details: ", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			ctx.Abort()
			return
		}

		ctx.Set("user", *user)
		ctx.Next()
	}
}
