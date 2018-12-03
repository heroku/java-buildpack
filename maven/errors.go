package maven

import (
	"fmt"
	"errors"
)

const (
	errorFmt = `
%s
  Caused by: %s

We're sorry this build is failing! If you can't find the issue in application code,
please submit a ticket so we can help: https://help.heroku.com/
`
)

func errorWithCause(message string, cause error) error {
	return errors.New(fmt.Sprintf(errorFmt, message, cause))
}

func failedToRunMaven(cause error) error {
	return errorWithCause("Failed to build app with Maven", cause)
}

func failedToDownloadSettings(cause error) error {
	return errorWithCause(fmt.Sprintf("Failed to download settings.xml from URL"), cause)
}

func failedToDownloadSettingsFromUrl(url string, cause error) error {
	return errorWithCause(fmt.Sprintf("Failed to download settings.xml from URL: %s", url), cause)
}