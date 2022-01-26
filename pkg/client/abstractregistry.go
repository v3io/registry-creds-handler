package client

import "github.com/nuclio/logger"

type abstractRegistry struct {
	logger logger.Logger
}

func newAbstractRegistry(loggerInstance logger.Logger) (*abstractRegistry, error) {
	return &abstractRegistry{
		logger: loggerInstance.GetChild("registry-generic"),
	}, nil
}
