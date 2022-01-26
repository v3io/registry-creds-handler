package pkg

import (
	"time"

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
	refreshRate   time.Duration
	registryKind  string
}

func NewHandler(logger logger.Logger,
	kubeClientSet *kubernetes.Clientset,
	ecr *registry.ECR,
	refreshRate int64,
	registryKind string) (*Handler, error) {

	return &Handler{
		logger:        logger.GetChild("handler"),
		kubeClientSet: kubeClientSet,
		ecr:           ecr,
		refreshRate:   time.Duration(refreshRate) * time.Minute,
		registryKind:  registryKind,
	}, nil
}

func (h *Handler) Run() error {
	h.logger.Info("Handler starting...")
	switch h.registryKind {
	case registry.ECRRegistryKind:
		err := h.createOrUpdateECRSecret()
		if err != nil {
			return errors.Wrap(err, "Failed to create or update ECR token secret")
		}

		// should never return
		util.SyncSecret(h.kubeClientSet, h.refreshRate, h.ecr.Namespace, h.ecr.SecretName, h.createOrUpdateECRSecret) // nolint: errcheck
	default:
		return errors.New("Received unsupported registry kind")
	}
	return errors.New("Handler exited unexpectedly")
}

func (h *Handler) createOrUpdateECRSecret() error {
	h.logger.InfoWith("Creating or updating ECR secret", "secretName", h.ecr.SecretName, "namespace", h.ecr.Namespace)

	// get token and generate secret
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

	// create or update secret
	_, err = util.GetSecret(h.kubeClientSet, h.ecr.Namespace, h.ecr.SecretName)
	if err != nil {
		h.logger.DebugWith("Secret not found, creating", "secretName", h.ecr.SecretName)

		err = util.CreateSecret(h.kubeClientSet, secretObj)
		if err != nil {
			return errors.Wrap(err, "Failed to create secret")
		}
		h.logger.InfoWith("Successfully created secret", "secretName", h.ecr.SecretName)

	} else {
		h.logger.DebugWith("Secret found, updating", "secretName", h.ecr.SecretName)

		err = util.UpdateSecret(h.kubeClientSet, secretObj)
		if err != nil {
			return errors.Wrap(err, "Failed to update secret")
		}
		h.logger.InfoWith("Successfully updated secret", "secretName", h.ecr.SecretName)
	}
	return nil
}
