package registrycredshandler

import (
	"context"
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
	registry      registry.Registry
	refreshRate   time.Duration
	registryKind  string
}

func NewHandler(logger logger.Logger,
	kubeClientSet *kubernetes.Clientset,
	registry registry.Registry,
	refreshRate int64,
	registryKind string) (*Handler, error) {

	return &Handler{
		logger:        logger.GetChild("handler"),
		kubeClientSet: kubeClientSet,
		registry:      registry,
		refreshRate:   time.Duration(refreshRate) * time.Minute,
		registryKind:  registryKind,
	}, nil
}

func (h *Handler) Start() error {
	h.logger.Info("Handler starting...")

	if err := h.createOrUpdateSecret(); err != nil {
		return errors.Wrap(err, "Failed to create or update secret")
	}

	// should never return
	//err := util.SyncSecret(h.kubeClientSet, h.refreshRate, h.registry.Namespace, h.registry.SecretName, h.createOrUpdateECRSecret)
	//if err != nil {
	//	return errors.Wrap(err, "Failed to sync secret")
	//}

	return errors.New("Handler exited unexpectedly")
}

// createOrUpdateSecret get token from registry, create or update secret with new token
func (h *Handler) createOrUpdateSecret() error {

	token, err := h.registry.GetAuthToken()
	if err != nil {
		return errors.Wrap(err, "Failed to get authorization token")
	}

	secret, err := util.GenerateSecretObj(token)
	if err != nil {
		return errors.Wrap(err, "Failed to generate secret object")
	}

	h.logger.DebugWith("Creating or updating secret", "SecretName", token.SecretName, "Namespace", token.SecretName)
	if err = util.CreateOrUpdateSecret(context.TODO(), h.kubeClientSet, "", secret); err != nil {
		return errors.Wrap(err, "Failed to create or update secret")
	}

	h.logger.Info("Secret created or updated successfully", "SecretName", token.SecretName, "Namespace", token.SecretName)
	return nil
}
