package registrycredshandler

import (
	"context"
	"time"

	"github.com/v3io/registry-creds-handler/pkg/common"
	"github.com/v3io/registry-creds-handler/pkg/registry"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	logger        logger.Logger
	kubeClientSet kubernetes.Interface
	registry      registry.Registry
	refreshRate   time.Duration
	registryKind  string
}

func NewHandler(logger logger.Logger,
	kubeClientSet kubernetes.Interface,
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
	h.logger.InfoWith("Handler starting...")

	// Create ctx, no need for cancel func
	ctx := context.Background()

	if err := h.createOrUpdateSecret(ctx); err != nil {
		return errors.Wrap(err, "Failed to create or update secret")
	}

	// spawn a goroutine for refreshing the secret
	go func() {

		h.logger.InfoWithCtx(ctx, "Starting secret refresher")
		if err := h.keepRefreshingSecret(ctx); err != nil {
			h.logger.ErrorWithCtx(ctx, "Failed and stopped refreshing secret", "err", err.Error())
			return
		}
		h.logger.WarnWithCtx(ctx, "Stopped refreshing secret")
	}()
	select {}
}

// keepRefreshingSecret will refresh the secret after every h.refreshRate until ctx is closed
func (h *Handler) keepRefreshingSecret(ctx context.Context) error {

	// Keep trying until we're timed out or got a result or got an error
	for {
		select {

		// Context was canceled, exit with error
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "Context was canceled, stopped refreshing secret")

		// Got a tick, time to refresh secret
		case <-time.After(h.refreshRate):
			if err := h.createOrUpdateSecret(ctx); err != nil {
				h.logger.WarnWithCtx(ctx, "Failed to refresh secret", "error", err.Error())
			}
		}
	}
}

// createOrUpdateSecret get token from registry, create or update secret with new token
func (h *Handler) createOrUpdateSecret(ctx context.Context) error {

	token, err := h.registry.GetAuthToken(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get authorization token")
	}

	secret, err := common.CompileRegistryAuthSecret(token)
	if err != nil {
		return errors.Wrap(err, "Failed to generate secret object")
	}

	h.logger.DebugWithCtx(ctx, "Creating or updating secret",
		"SecretName", token.SecretName,
		"Namespace", token.Namespace)

	if err := common.CreateOrUpdateSecret(ctx, h.kubeClientSet, secret); err != nil {
		return errors.Wrap(err, "Failed to create or update secret")
	}

	h.logger.InfoWithCtx(ctx, "Secret created or updated successfully",
		"SecretName", token.SecretName,
		"Namespace", token.Namespace)
	return nil
}
