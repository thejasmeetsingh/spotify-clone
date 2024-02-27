package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	log "github.com/sirupsen/logrus"
	"github.com/thejasmeetsingh/spotify-clone/src/user_service/database"
	"github.com/thejasmeetsingh/spotify-clone/src/user_service/utils"
	"github.com/thejasmeetsingh/spotify-clone/src/user_service/validators"
)

func SignUp(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		type Parameters struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		var params Parameters
		err := ctx.ShouldBindJSON(&params)

		if err != nil {
			log.Errorln("Error caught while parsing signup request data: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Error while parsing the request data"})
			return
		}

		// Make the password case sensitive
		email := strings.ToLower(params.Email)

		// Validate Password
		err = validators.PasswordValidator(params.Password, email)
		if err != nil {
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Generate hashed password
		hashedPassword, err := utils.GetHashedPassword(params.Password)

		if err != nil {
			log.Errorln("Error caught while generating hashed password: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		// Check if user exists with the given email address
		_, err = database.GetUserByEmailDB(dbCfg, ctx, email)
		if err == nil {
			ctx.SecureJSON(http.StatusForbidden, gin.H{"message": "User with this email address already exists"})
			return
		}

		// Begin DB transaction
		tx, err := dbCfg.DB.Begin(ctx)
		if err != nil {
			log.Fatalln("Error caught while starting a transaction: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}
		defer tx.Rollback(ctx)
		qtx := dbCfg.Queries.WithTx(tx)

		// Create user account
		currentTime := time.Now().UTC()

		dbUser, err := qtx.CreateUser(ctx, database.CreateUserParams{
			ID: uuid.New(),
			CreatedAt: pgtype.Timestamp{
				Time:  currentTime,
				Valid: true,
			},
			ModifiedAt: pgtype.Timestamp{
				Time:  currentTime,
				Valid: true,
			},
			Email:    email,
			Password: hashedPassword,
		})

		if err != nil {
			log.Errorln("Error caught while creating a user in DB: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		// Generate auth tokens for the user
		tokens, err := utils.GenerateTokens(dbUser.ID.String())

		if err != nil {
			log.Errorln("Error caught while generating auth tokens during signup: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		// Commit the transaction
		err = tx.Commit(ctx)
		if err != nil {
			log.Fatalln("Error caught while closing a transaction: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		ctx.SecureJSON(http.StatusCreated, gin.H{"message": "Account created successfully!", "data": tokens})
	}
}

// Login API
func Login(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		type Parameters struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		var params Parameters
		err := ctx.ShouldBindJSON(&params)

		if err != nil {
			log.Errorln("Error caught while parsing login request data: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Error while parsing the request data"})
			return
		}

		// Check weather the user exists with the given email or not
		user, err := database.GetUserByEmailDB(dbCfg, ctx, strings.ToLower(params.Email))
		if err != nil {
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "User does not exists, Please check your credentials"})
			return
		}

		// Check the given password with hashed password stored in DB
		match, err := utils.CheckPasswordValid(params.Password, user.Password)
		if err != nil {
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid password"})
			return
		} else if !match {
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid password"})
			return
		}

		// Generate auth tokens for the user
		tokens, err := utils.GenerateTokens(user.ID.String())

		if err != nil {
			log.Errorln("Error caught while generating auth tokens during login: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		ctx.SecureJSON(http.StatusOK, gin.H{"message": "Logged in Successfully!", "data": tokens})
	}
}

// Refresh Token API
//
// Generate new tokens if the given refresh token is valid
func RefreshAccessToken(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		type Parameters struct {
			RefreshToken string `json:"refresh_token"`
		}

		var params Parameters
		err := ctx.ShouldBindJSON(&params)

		if err != nil {
			log.Errorln("Error caught while parsing refresh token request data: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Error while parsing the request data"})
			return
		}

		tokens, err := utils.ReIssueAccessToken(params.RefreshToken)
		if err != nil {
			log.Errorln("Error caught while re-issuing auth tokens: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Error while issuing new tokens"})
			return
		}

		ctx.SecureJSON(http.StatusOK, gin.H{"message": "Tokens re-issued Successfully!", "data": tokens})
	}
}
