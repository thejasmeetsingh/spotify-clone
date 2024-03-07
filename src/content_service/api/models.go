package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thejasmeetsingh/spotify-clone/src/content_service/database"
	"github.com/thejasmeetsingh/spotify-clone/src/content_service/internal"
)

type Content struct {
	ID          uuid.UUID     `json:"id"`
	CreatedAt   time.Time     `json:"created_at"`
	ModifiedAt  time.Time     `json:"modified_at"`
	User        internal.User `json:"user"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Type        string        `json:"type"`
	Url         string        `json:"url"`
}

type ContentList struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
}

func databaseContentToContent(content *database.Content, user internal.User) Content {
	return Content{
		ID:          content.ID,
		CreatedAt:   content.CreatedAt.Time,
		ModifiedAt:  content.ModifiedAt.Time,
		Title:       content.Title,
		Description: content.Description,
		Type:        string(content.Type),
		User:        user,
		Url:         "",
	}
}

func databaseContentListToContentList(dbContentList []interface{}) ([]ContentList, error) {
	var contentList []ContentList

	for _, dbContent := range dbContentList {
		value, ok := dbContent.(ContentList)
		if !ok {
			return contentList, fmt.Errorf("error caught while converting the content list")
		}
		contentList = append(contentList, value)
	}
	return contentList, nil
}
