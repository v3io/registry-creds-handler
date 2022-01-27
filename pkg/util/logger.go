package util

import (
	"io"

	"github.com/nuclio/errors"
	"github.com/nuclio/loggerus"
	"github.com/sirupsen/logrus"
)

func CreateLogger(name string, verbose bool, output io.Writer, logsFormat string) (*loggerus.Loggerus, error) {

	var logLevel logrus.Level
	if verbose {
		logLevel = logrus.DebugLevel
	} else {
		logLevel = logrus.InfoLevel
	}

	var logger *loggerus.Loggerus
	var err error

	// TODO: add redactor
	switch logsFormat {
	case "json":
		logger, err = loggerus.NewJSONLoggerus(name, logLevel, output)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create new json logger")
		}
	case "humanreadable":
		logger, err = loggerus.NewTextLoggerus(name, logLevel, output, true, false)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create new text logger")
		}
	default:
		return nil, errors.Wrapf(err, "Failed to create logger, received unexpected log format: %s", logsFormat)
	}

	return logger, nil
}
