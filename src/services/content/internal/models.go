package internal

import (
	"encoding/json"

	"github.com/google/uuid"
)

type User struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

func UserToByte(user User) ([]byte, error) {
	return json.Marshal(user)
}

func ByteToUser(userByte []byte) (*User, error) {
	var user User

	err := json.Unmarshal(userByte, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
