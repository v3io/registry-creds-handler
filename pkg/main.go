package pkg

import (
	"github.com/v3io/registry-creds-handler/pkg/registry"
	"github.com/v3io/registry-creds-handler/pkg/util"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	logger        logger.Logger
	kubeClientSet *kubernetes.Clientset
	ecr           *registry.ECR
	refreshRate   int
	registryKind  string
}

func NewHandler(logger logger.Logger,
	kubeClientSet *kubernetes.Clientset,
	ecr *registry.ECR,
	refreshRate int,
	registryKind string) (*Handler, error) {

	return &Handler{
		logger:        logger.GetChild("handler"),
		kubeClientSet: kubeClientSet,
		ecr:           ecr,
		refreshRate:   refreshRate,
		registryKind:  registryKind,
	}, nil
}

func (h *Handler) Run() error {
	h.logger.Info("Handler starting...")
	switch h.registryKind {
	case registry.ECRRegistryKind:
		err := h.createECRTokenSecret()
		if err != nil {
			return errors.Wrap(err, "Failed to create ECR token secret")
		}
	default:
		return errors.New("Received unsupported registry kind")
	}
	return errors.New("Handler exited unexpectedly")
}

func (h *Handler) createECRTokenSecret() error {
	h.logger.Info("Creating ECR token secret")

	secret, err := util.GetSecret(h.kubeClientSet, h.ecr.Namespace, h.ecr.SecretName)
	if err == nil && secret != nil {
		return errors.Wrapf(err, "Secret `%s` already exists", h.ecr.SecretName)
	}

	accessToken, err := h.ecr.GetAuthorizationToken()
	if err != nil {
		return errors.Wrap(err, "Failed to get authorization token")
	}

	token := registry.Token{
		SecretName:  h.ecr.SecretName,
		AccessToken: accessToken,
		Endpoints:   h.ecr.Endpoints,
	}
	secretObj := util.GenerateSecretObj(token)
	_, err = util.CreateSecret(h.kubeClientSet, secretObj)
	if err != nil {
		return errors.Wrap(err, "Failed to create secret")
	}

	return nil
}
