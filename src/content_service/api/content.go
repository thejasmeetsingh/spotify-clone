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
)

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

func getUserID(ctx *gin.Context) (uuid.UUID, error) {
	userIDStr, isExists := ctx.Get("userID")
	if !isExists {
		return uuid.Nil, fmt.Errorf("authentication failed")
	}

	userID, err := uuid.Parse(fmt.Sprintf("%v", userIDStr))
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func getContentList(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		offset := getOffset(ctx)

		dbContentList, err := database.GetContentListDB(dbCfg, ctx, database.GetContentListParams{
			Limit:  10,
			Offset: offset,
		})

		if err != nil {
			log.Fatalln("error caught while fetching content list: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		contentList, err := databaseContentListToContentList(dbContentList)
		if err != nil {
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}
		if len(contentList) == 0 {
			ctx.JSON(http.StatusOK, gin.H{"results": []string{}})
			return
		}
		ctx.SecureJSON(http.StatusOK, gin.H{"results": contentList})
	}
}

func getUserContentList(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, err := getUserID(ctx)
		if err != nil {
			log.Fatalln(err)
			ctx.SecureJSON(http.StatusBadRequest, gin.H{"message": "Something went wrong"})
			return
		}

		offset := getOffset(ctx)

		dbContentUserList, err := database.GetUserContentDB(dbCfg, ctx, database.GetUserContentParams{
			UserID: userID,
			Limit:  10,
			Offset: offset,
		})

		if err != nil {
			log.Fatalln("error caught while fetching user content list: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		userContentList, err := databaseContentListToContentList(dbContentUserList)
		if err != nil {
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}
		if len(userContentList) == 0 {
			ctx.JSON(http.StatusOK, gin.H{"results": []string{}})
			return
		}
		ctx.SecureJSON(http.StatusOK, gin.H{"results": userContentList})
	}
}

func getContentDetail(dbCfg *database.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		contentIDStr := ctx.Param("id")
		contentID, err := uuid.Parse(contentIDStr)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid content ID format"})
			return
		}

		dbContent, err := database.GetContentDetailDB(dbCfg, ctx, contentID)
		if err != nil {
			log.Fatalln("error caught while fetching content detail: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		ctx.SecureJSON(http.StatusOK, gin.H{"data": databaseContentToContent(dbContent)})
	}
}

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

		userID, err := getUserID(ctx)
		if err != nil {
			log.Fatalln(err)
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
			UserID:      userID,
			Title:       params.Title,
			Description: params.Description,
			Type:        database.ContentType(params.Type),
		})

		if err != nil {
			log.Fatalln("error caught while adding content details to DB: ", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
			return
		}

		ctx.SecureJSON(http.StatusCreated, gin.H{"data": databaseContentToContent(dbContent)})
	}
}

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
		log.Fatalln("error caught while loading aws config: ", err)
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
		log.Fatalln("error caught while generating pre-sign upload URL: ", err)
		ctx.SecureJSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
		return
	}

	resData := map[string]string{
		"url": res.URL,
		"key": s3_key,
	}

	ctx.SecureJSON(http.StatusOK, gin.H{"data": resData})
}
