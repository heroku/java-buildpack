package jdk

import (
	"errors"
	"fmt"
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

func invalidJdkVersion(version string, url string) error {
	return errorWithCause(fmt.Sprintf("Invalid JDK version: %s", version), errors.New(fmt.Sprintf("Failed to reach %s", url)))
}
