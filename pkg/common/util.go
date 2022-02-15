package common

import (
	"encoding/base64"
	"strings"

	"github.com/nuclio/errors"
)

func GetFirstNonEmptyString(strings []string) string {
	for _, s := range strings {
		if s != "" {
			return s
		}
	}
	return ""
}

func ParseAuth(auth string) (string, string, error) {
	decodedAuth, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return "", "", errors.Wrap(err, "Failed to decode auth")
	}

	splitAuth := strings.Split(string(decodedAuth), ":")
	if len(splitAuth) < 2 {
		return "", "", errors.Wrap(err, "Failed to split auth")
	}

	// username, password
	return splitAuth[0], splitAuth[1], nil
}
