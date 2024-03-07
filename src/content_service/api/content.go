package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	log "github.com/sirupsen/logrus"
	"github.com/thejasmeetsingh/spotify-clone/src/content_service/database"
	"github.com/thejasmeetsingh/spotify-clone/src/content_service/internal"
)

// Parse the offset from query params
func getOffset(ctx *gin.Context) int32 {
	offsetStr := ctx.Query("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}
	return int32(offset)
}

// Get user object from the context
func getUser(ctx *gin.Context) (internal.User, error) {
	value, exists := ctx.Get("user")
	if !exists {
		return internal.User{}, fmt.Errorf("authentication required")
	}

	user, ok := value.(internal.User)
	if !ok {
		return internal.User{}, fmt.Errorf("invalid user")
	}

	return user, nil
}

// API for getting list of content present on the system
// Non-auth API: Anyone can view contents
func getContentList(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		offset := getOffset(ctx)

		// Fetch content list from DB
		dbContentList, err := database.GetContentListDB(dbCfg, ctx, database.GetContentListParams{
			Limit:  10,
			Offset: offset,
		})

		if err != nil {
			log.Errorln("error caught while fetching content list: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		// Return an empty array if db contents list is empty
		if len(dbContentList) == 0 {
			ctx.JSON(http.StatusOK, gin.H{"results": []string{}})
			return
		}

		// Parse DB content list with appropriate key names
		contentList, err := databaseContentListToContentList(dbContentList)
		if err != nil {
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		ctx.SecureJSON(http.StatusOK, gin.H{"results": contentList})
	}
}

// API for getting contents added by current user
func getUserContentList(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := getUser(ctx)
		if err != nil {
			log.Errorln(err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Something went wrong"})
			return
		}

		offset := getOffset(ctx)

		// Fetch user contents from DB
		dbContentUserList, err := database.GetUserContentDB(dbCfg, ctx, database.GetUserContentParams{
			UserID: user.ID,
			Limit:  10,
			Offset: offset,
		})

		if err != nil {
			log.Errorln("error caught while fetching user content list: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		// Return an empty array if user contents list is empty
		if len(dbContentUserList) == 0 {
			ctx.JSON(http.StatusOK, gin.H{"results": []string{}})
			return
		}

		// Parse DB content list with appropriate key names
		userContentList, err := databaseUserContentListToContentList(dbContentUserList)
		if err != nil {
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		ctx.SecureJSON(http.StatusOK, gin.H{"results": userContentList})
	}
}

// API for getting content detail
// Non-auth API: Anyone can view the content details
func getContentDetail(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		contentID, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid content ID format"})
			return
		}

		dbContent, err := database.GetContentDetailDB(dbCfg, ctx, contentID)
		if err != nil {
			log.Errorln("error caught while fetching content detail: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		ctx.SecureJSON(http.StatusOK, gin.H{"data": databaseContentToContent(dbContent)})
	}
}

// API for adding a content into the DB
func addContent(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		type Parameters struct {
			Title       string `json:"title" binding:"required"`
			Description string `json:"description" binding:"required"`
			Type        string `json:"type" binding:"required"`
		}
		var params Parameters

		err := ctx.ShouldBindJSON(&params)
		if err != nil {
			log.Errorln("error while parsing request data: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
			return
		}

		user, err := getUser(ctx)
		if err != nil {
			log.Errorln(err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		dbContent, err := database.AddContentDB(dbCfg, ctx, database.AddContentParams{
			ID: uuid.New(),
			CreatedAt: pgtype.Timestamp{
				Time:  time.Now().UTC(),
				Valid: true,
			},
			ModifiedAt: pgtype.Timestamp{
				Time:  time.Now().UTC(),
				Valid: true,
			},
			UserID:      user.ID,
			Title:       params.Title,
			Description: params.Description,
			Type:        database.ContentType(params.Type),
		})

		if err != nil {
			log.Errorln("error caught while adding content details to DB: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		ctx.SecureJSON(http.StatusCreated, gin.H{"data": databaseContentToContent(dbContent)})
	}
}

func updateContent(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		type Parameters struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Type        string `json:"type"`
		}
		var params Parameters

		// Parse request data
		err := ctx.ShouldBindJSON(&params)

		if err != nil {
			log.Errorln("error while parsing request data: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
			return
		}

		// Parse content ID passed in request path
		contentID, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid content ID"})
			return
		}

		// Fetch content record from DB
		dbContent, err := database.GetContentDetailDB(dbCfg, ctx, contentID)
		if err != nil {
			log.Errorln("error caught while fetching content detail: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		// Pre-fill empty values
		// So that no empty values gets saved in DB
		if params.Title == "" {
			params.Title = dbContent.Title
		}

		if params.Description == "" {
			params.Description = dbContent.Description
		}

		if params.Type == "" {
			params.Type = string(dbContent.Type)
		}

		// Update content detail in DB
		dbContent, err = database.UpdateContentDetailDB(dbCfg, ctx, database.UpdateContentDetailsParams{
			ID:          contentID,
			Title:       params.Title,
			Description: params.Description,
			Type:        database.ContentType(params.Type),
			ModifiedAt: pgtype.Timestamp{
				Time:  time.Now().UTC(),
				Valid: true,
			},
		})
		if err != nil {
			log.Errorln("error caught while updating content detail: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		ctx.SecureJSON(http.StatusOK, gin.H{"data": databaseContentToContent(dbContent)})
	}
}

func updateContentS3Key(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		type Parameters struct {
			Key string `json:"key" binding:"required"`
		}
		var params Parameters

		// Parse request data
		err := ctx.ShouldBindJSON(&params)
		if err != nil {
			log.Errorln("error while parsing request data: ", err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
			return
		}

		// Parse content ID passed in request path
		contentID, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid content ID"})
			return
		}

		// Update S3 key in DB
		if err = database.UpdateContentS3KeyDB(dbCfg, ctx, database.UpdateS3KeyParams{
			ID: contentID,
			S3Key: pgtype.Text{
				String: params.Key,
				Valid:  true,
			},
			ModifiedAt: pgtype.Timestamp{
				Time:  time.Now().UTC(),
				Valid: true,
			},
		}); err != nil {
			log.Errorln("error caught while updating content s3 key: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		// TODO: gRPC to conversion service for processing the media file

		ctx.SecureJSON(http.StatusOK, gin.H{"message": "Key updated successfully"})
	}
}

func deleteContent(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Parse content ID passed in request path
		contentID, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid content ID"})
			return
		}

		user, err := getUser(ctx)
		if err != nil {
			log.Errorln(err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Something went wrong"})
			return
		}

		// Delete content from DB
		if err = database.DeleteContentDB(dbCfg, ctx, database.DeleteContentParams{
			ID:     contentID,
			UserID: user.ID,
		}); err != nil {
			log.Errorln("error caught while deleting content from DB: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		ctx.SecureJSON(http.StatusOK, gin.H{"message": "Content deleted successfully"})
	}
}

// API for getting pre-signed URL for file upload
func getPresignedURL(ctx *gin.Context) {
	type Parameters struct {
		ContentID   string `json:"content_id" binding:"required"`
		FileName    string `json:"filename" binding:"required"`
		IsAudioFile bool   `json:"is_audio_file" binding:"required"`
	}
	var params Parameters
	err := ctx.ShouldBindJSON(&params)

	if err != nil {
		log.Errorln("error while parsing request data: ", err)
		ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
		return
	}

	// Load default AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Errorln("error caught while loading aws config: ", err)
		ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
		return
	}

	// Create s3 client
	client := s3.NewPresignClient(s3.NewFromConfig(cfg))
	s3_key := getUniqueFilename(params.ContentID, params.FileName, params.IsAudioFile)

	// Generate pre-signed URL for file upload
	res, err := client.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("AWS_REGION")),
		Key:    aws.String(s3_key),
	})

	if err != nil {
		log.Errorln("error caught while generating pre-sign upload URL: ", err)
		ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
		return
	}

	// Prepare response data
	resData := map[string]string{
		"url": res.URL,
		"key": s3_key,
	}

	ctx.SecureJSON(http.StatusOK, gin.H{"data": resData})
}
