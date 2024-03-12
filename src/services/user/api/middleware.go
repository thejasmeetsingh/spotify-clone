package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thejasmeetsingh/spotify-clone/src/services/user/database"
	"github.com/thejasmeetsingh/spotify-clone/src/services/user/utils"
)

// Validate the request by checking weather or not they have the valid JWT access token or not
//
// Token format: Bearer <TOKEN>
func JWTAuth(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		headerAuthToken := ctx.GetHeader("Authorization")

		if headerAuthToken == "" {
			ctx.SecureJSON(http.StatusForbidden, gin.H{"message": "Authentication required"})
			ctx.Abort()
			return
		}

		// Split the token string
		authToken := strings.Split(headerAuthToken, " ")

		// Validate the token string
		if len(authToken) != 2 || authToken[0] != "Bearer" {
			ctx.SecureJSON(http.StatusForbidden, gin.H{"message": "Invalid authentication format"})
			ctx.Abort()
			return
		}

		// Verify the token and get the encoded payload which is the userID string
		claims, err := utils.VerifyToken(authToken[1])
		if err != nil {
			ctx.SecureJSON(http.StatusForbidden, gin.H{"message": "Invalid authentication token"})
			ctx.Abort()
			return
		}

		// Check the validity of the token
		if !time.Unix(claims.ExpiresAt.Unix(), 0).After(time.Now()) {
			ctx.SecureJSON(http.StatusForbidden, gin.H{"message": "Authentication token is expired"})
			ctx.Abort()
			return
		}

		// Convert the userID string to UUID
		userID, err := uuid.Parse(claims.Data)
		if err != nil {
			ctx.SecureJSON(http.StatusForbidden, gin.H{"message": "Invalid authentication token"})
			ctx.Abort()
			return
		}

		// Fetch user by ID from DB
		dbUser, err := database.GetUserByIDFromDB(dbCfg, ctx, userID)
		if err != nil {
			ctx.SecureJSON(http.StatusForbidden, gin.H{"message": "Something went wrong"})
			ctx.Abort()
			return
		}

		ctx.Set("user", databaseUserToUser(dbUser))

		// Further call the given handler and send the user instance as well
		ctx.Next()
	}
}
