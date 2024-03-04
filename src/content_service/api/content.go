package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
