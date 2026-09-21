package fleet

import "strings"

func splitGatewayPath(requestPath, escapedPath string) (string, string, string, bool) {
	trimmedPath := strings.TrimPrefix(requestPath, "/")
	id, remainingPath, hasRemainingPath := strings.Cut(trimmedPath, "/")
	if id == "" {
		return "", "", "", false
	}
	if hasRemainingPath {
		remainingPath = "/" + remainingPath
	} else {
		remainingPath = "/"
	}

	trimmedEscapedPath := strings.TrimPrefix(escapedPath, "/")
	_, remainingEscapedPath, hasRemainingEscapedPath := strings.Cut(trimmedEscapedPath, "/")
	if hasRemainingEscapedPath {
		remainingEscapedPath = "/" + remainingEscapedPath
	} else {
		remainingEscapedPath = ""
	}

	return id, remainingPath, remainingEscapedPath, true
}
