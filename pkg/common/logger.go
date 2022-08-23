/*
Copyright 2022 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/
package common

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
