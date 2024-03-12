package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	log "github.com/sirupsen/logrus"
	"github.com/thejasmeetsingh/spotify-clone/src/services/user/database"
	"github.com/thejasmeetsingh/spotify-clone/src/services/user/utils"
	"github.com/thejasmeetsingh/spotify-clone/src/services/user/validators"
)

// Common function for getting user object from the context, For all handlers in this module
func getUserFromCtx(ctx *gin.Context) (User, error) {
	value, exists := ctx.Get("user")

	if !exists {
		return User{}, fmt.Errorf("authentication required")
	}

	user, ok := value.(User)

	if !ok {
		return User{}, fmt.Errorf("invalid user")
	}

	return user, nil
}

// Fetch user profile details
func getUserProfile(ctx *gin.Context) {
	user, err := getUserFromCtx(ctx)
	if err != nil {
		ctx.SecureJSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}
	ctx.SecureJSON(http.StatusOK, gin.H{"data": user})
}

// Update user profile details
func updateUserProfile(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := getUserFromCtx(ctx)
		if err != nil {
			ctx.SecureJSON(http.StatusForbidden, gin.H{"message": err.Error()})
			return
		}

		type Parameters struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		var params Parameters

		if err = ctx.ShouldBindJSON(&params); err != nil {
			log.Errorln("Error caught while parsing update user detail API request data: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Error while parsing the request data"})
			return
		}

		// check if request body is empty
		if params.Name == "" && params.Email == "" {
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
			return
		}

		// Pre-fill name and email address with existing data from DB
		if params.Name == "" {
			params.Name = user.Name
		}

		isEmailChanged := true
		if params.Email == "" {
			params.Email = user.Email
			isEmailChanged = false
		}

		email := strings.ToLower(params.Email)

		// Validate the given email address
		if !validators.EmailValidator(params.Email) {
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid email address"})
			return
		}

		// Check if user any exists with the new email address
		if isEmailChanged && email != user.Email {
			if _, err = database.GetUserByEmailDB(dbCfg, ctx, email); err == nil {
				ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "User with this email address already exists"})
				return
			}
		}

		// Update the profile details
		dbUser, err := database.UpdateUserDetailDB(dbCfg, ctx, database.UpdateUserDetailsParams{
			Name: pgtype.Text{
				String: params.Name,
				Valid:  true,
			},
			Email: email,
			ModifiedAt: pgtype.Timestamp{
				Time:  time.Now().UTC(),
				Valid: true,
			},
			ID: user.ID,
		})

		if err != nil {
			log.Fatalln("error while updating user details: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		user = databaseUserToUser(dbUser)
		ctx.SecureJSON(http.StatusOK, gin.H{"message": "Profile details updated successfully!", "data": user})
	}
}

// Delete user profile
func deleteUserProfile(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := getUserFromCtx(ctx)
		if err != nil {
			ctx.SecureJSON(http.StatusForbidden, gin.H{"message": err.Error()})
			return
		}

		if err = database.DeleteUserDB(dbCfg, ctx, user.ID); err != nil {
			ctx.SecureJSON(http.StatusForbidden, gin.H{"message": "Something went wrong"})
			return
		}
		ctx.SecureJSON(http.StatusOK, gin.H{"message": "Profile deleted successfully!"})
	}
}

// Change password API for authenticated user
func changePassword(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := getUserFromCtx(ctx)
		if err != nil {
			ctx.SecureJSON(http.StatusForbidden, gin.H{"message": err.Error()})
			return
		}

		type Parameters struct {
			OldPassword string `json:"old_password" binding:"required"`
			NewPassword string `json:"new_password" binding:"required"`
		}

		var params Parameters
		if err = ctx.ShouldBindJSON(&params); err != nil {
			log.Errorln("Error caught while parsing change password API request data: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Error while parsing the request data"})
			return
		}

		// Fetch user by ID from DB
		dbUser, err := database.GetUserByIDFromDB(dbCfg, ctx, user.ID)
		if err != nil {
			log.Fatalln("error while getting user details by ID: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		// Check weather or not old password is correct or not
		match, err := utils.CheckPasswordValid(params.OldPassword, dbUser.Password)
		if err != nil {
			log.Errorln("Error caught while checking current password: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid old password, Please try again."})
			return
		} else if !match {
			log.Errorln("Error caught while checking current password: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid old password, Please try again."})
			return
		}

		if params.OldPassword == params.NewPassword {
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "New password should not be same as old password"})
			return
		}

		// Validate the new password
		if err = validators.PasswordValidator(params.NewPassword, dbUser.Email); err != nil {
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Generate the new hashed password
		hashedPassword, err := utils.GetHashedPassword(params.NewPassword)

		if err != nil {
			log.Errorln("Error caught while generating hashed password: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Something went wrong"})
			return
		}

		// Update the password
		err = database.UpdateUserPasswordDB(dbCfg, ctx, database.UpdateUserPasswordParams{
			Password: hashedPassword,
			ModifiedAt: pgtype.Timestamp{
				Time:  time.Now().UTC(),
				Valid: true,
			},
			ID: dbUser.ID,
		})

		if err != nil {
			log.Fatalln("error while updating password: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		ctx.SecureJSON(http.StatusOK, gin.H{"message": "Password changed successfully!"})
	}
}
