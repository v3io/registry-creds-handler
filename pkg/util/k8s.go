package util

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/v3io/registry-creds-handler/pkg/registry"

	"github.com/nuclio/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type DockerConfigJSON struct {
	Auths map[string]RegistryAuth `json:"auths,omitempty"`
}

type RegistryAuth struct {
	Auth string `json:"auth"`
}

func GetClientConfig(kubeConfigPath string) (*rest.Config, error) {
	if kubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}

	return rest.InClusterConfig()
}

func NewKubeClientSet(kubeConfigPath string) (*kubernetes.Clientset, error) {

	cfg, err := GetClientConfig(kubeConfigPath)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to get kube client config")
	}

	clientSet, err := kubernetes.NewForConfig(cfg)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to get kube clientset")
	}

	return clientSet, nil
}

// GetSecret get a secret
func GetSecret(ctx context.Context,
	kubeClient *kubernetes.Clientset,
	namespace string,
	secretName string) (*v1.Secret, error) {
	secret, err := kubeClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get secret")
	}

	return secret, nil
}

// CreateSecret creates a secret
func CreateSecret(ctx context.Context,
	kubeClient *kubernetes.Clientset,
	secret *v1.Secret) error {
	_, err := kubeClient.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metav1.CreateOptions{})

	if err != nil {
		return errors.Wrapf(err, "Failed to create secret: %s", secret.Name)
	}

	return nil
}

// UpdateSecret updates a secret
func UpdateSecret(ctx context.Context,
	kubeClient *kubernetes.Clientset,
	secret *v1.Secret) error {
	_, err := kubeClient.CoreV1().Secrets(secret.Namespace).Update(ctx, secret, metav1.UpdateOptions{})

	if err != nil {
		return errors.Wrapf(err, "Failed to update secret: %s", secret.Name)
	}

	return nil
}

func CreateOrUpdateSecret(ctx context.Context,
	kubeClient *kubernetes.Clientset,
	namespace string,
	secret *v1.Secret) error {
	_, err := GetSecret(ctx, kubeClient, namespace, secret.Name)
	if err != nil {
		if err = CreateSecret(ctx, kubeClient, secret); err != nil {
			return errors.Wrap(err, "Failed to create secret")
		}

	} else {
		if err = UpdateSecret(ctx, kubeClient, secret); err != nil {
			return errors.Wrap(err, "Failed to update secret")
		}
	}
	return nil
}

// GenerateSecretObj creates a secret object with given access token and docker config for imagePullSecrets if possible
func GenerateSecretObj(token *registry.Token) (*v1.Secret, error) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: token.SecretName,
		},
	}

	// If possible, add auth to docker config
	if token.RegistryUri != "" {

		// { "auths": { "<registry-endpoint>": { "auth": "..." }}}
		auths := map[string]RegistryAuth{}
		auths[token.RegistryUri] = RegistryAuth{
			Auth: *token.AccessToken,
		}

		configJSON, err := json.Marshal(DockerConfigJSON{Auths: auths})
		if err != nil {
			return nil, errors.Wrap(err, "Failed to marshal docker config json")
		}

		secret.Data = map[string][]byte{
			".dockerconfigjson": configJSON,
			"ACCESS_TOKEN":      []byte(*token.AccessToken),
		}
		secret.Type = "kubernetes.io/dockerconfigjson"

	} else {
		secret.Data = map[string][]byte{
			"ACCESS_TOKEN": []byte(*token.AccessToken),
		}
	}

	return secret, nil
}

// SyncSecret watches a secret in a specific namespace and handles the secret every interval
func SyncSecret(kubeClient *kubernetes.Clientset,
	interval time.Duration,
	namespace string,
	secretName string,
	handler func() error) error {

	// once closed, watching the secret will seize
	stopCh := make(chan struct{})

	// Extend the selector to include specific secret to monitor
	selector, err := fields.ParseSelector("metadata.name=" + secretName)
	if err != nil {
		return errors.Wrap(err, "Failed to create secret selector")
	}

	lw := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "secrets", namespace, selector)
	_, c := cache.NewInformer(
		lw,
		&v1.Secret{},
		interval,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if err := handler(); err != nil {
					fmt.Println(err)
				}
			},
			UpdateFunc: func(_ interface{}, obj interface{}) {
				if err := handler(); err != nil {
					fmt.Println(err)
				}
			},
		},
	)
	c.Run(stopCh)
	return nil
}
