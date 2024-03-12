// Custom user model for converting the raw data into a desired JSON data
//
// With keys as formatted as snake_case rather than TitleCase

package api

import (
	"github.com/google/uuid"
	"github.com/thejasmeetsingh/spotify-clone/src/services/user/database"
)

type User struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

func databaseUserToUser(dbUser *database.User) User {
	return User{
		ID:    dbUser.ID,
		Name:  dbUser.Name.String,
		Email: dbUser.Email,
	}
}
