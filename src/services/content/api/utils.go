package api

import "strings"

func getUniqueFilename(contentID, srcFilename string, isAudioFile bool) string {
	filename := contentID + "." + strings.Split(srcFilename, ".")[1]

	if isAudioFile {
		return "audio/" + filename
	}
	return "video/" + filename
}
