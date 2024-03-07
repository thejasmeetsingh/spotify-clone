package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/thejasmeetsingh/spotify-clone/src/content_service/database"
)

type Content struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Url         string    `json:"url"`
}

type ContentList struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
}

func databaseContentToContent(content *database.Content) Content {
	return Content{
		ID:          content.ID,
		CreatedAt:   content.CreatedAt.Time,
		ModifiedAt:  content.ModifiedAt.Time,
		Title:       content.Title,
		Description: content.Description,
		Type:        string(content.Type),
		Url:         "",
	}
}

func databaseContentListToContentList(dbContentList []database.GetContentListRow) ([]ContentList, error) {
	var contentList []ContentList

	for _, dbContent := range dbContentList {
		contentList = append(contentList, ContentList{
			ID:          dbContent.ID,
			CreatedAt:   dbContent.CreatedAt.Time,
			Title:       dbContent.Title,
			Description: dbContent.Description,
			Type:        string(dbContent.Type),
		})
	}

	return contentList, nil
}

func databaseUserContentListToContentList(dbContentList []database.GetUserContentRow) ([]ContentList, error) {
	var contentList []ContentList

	for _, dbContent := range dbContentList {
		contentList = append(contentList, ContentList{
			ID:          dbContent.ID,
			CreatedAt:   dbContent.CreatedAt.Time,
			Title:       dbContent.Title,
			Description: dbContent.Description,
			Type:        string(dbContent.Type),
		})
	}

	return contentList, nil
}
