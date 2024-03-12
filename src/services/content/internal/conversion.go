package internal

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	log "github.com/sirupsen/logrus"
	"github.com/thejasmeetsingh/spotify-clone/src/services/content/database"
)

func UpdateContentS3Key(
	dbCfg *database.Config,
	ctx context.Context,
	params database.UpdateS3KeyParams,
	key string,
	isAudioFile bool,
) {
	// Process the media file and retrieve the new s3 key
	grpcResponse, err := processContentMedia(key, isAudioFile)
	if err != nil {
		log.Errorln("error caught in conversion gRPC response: ", err)
		return
	}

	// Update s3 key
	params.S3Key = pgtype.Text{
		String: grpcResponse.GetKey(),
		Valid:  true,
	}

	// Save the updated key in DB
	if err = database.UpdateContentS3KeyDB(dbCfg, ctx, params); err != nil {
		log.Errorln("error caught while updating content s3 key: ", err)
		return
	}

	log.Infoln("Content s3 key updated successfully")
}
