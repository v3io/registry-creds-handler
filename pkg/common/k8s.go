package common

import (
	"context"
	"encoding/json"

	"github.com/v3io/registry-creds-handler/pkg/registry"

	"github.com/nuclio/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

func NewKubeClientSet(kubeConfigPath string) (kubernetes.Interface, error) {

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
	kubeClient kubernetes.Interface,
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
	kubeClient kubernetes.Interface,
	secret *v1.Secret) error {

	if _, err := kubeClient.CoreV1().Secrets(secret.Namespace).Create(ctx,
		secret,
		metav1.CreateOptions{}); err != nil {
		return errors.Wrapf(err, "Failed to create secret: %s", secret.Name)
	}

	return nil
}

// UpdateSecret updates a secret
func UpdateSecret(ctx context.Context,
	kubeClient kubernetes.Interface,
	secret *v1.Secret) error {

	if _, err := kubeClient.CoreV1().Secrets(secret.Namespace).Update(ctx,
		secret,
		metav1.UpdateOptions{}); err != nil {
		return errors.Wrapf(err, "Failed to update secret: %s", secret.Name)
	}

	return nil
}

func CreateOrUpdateSecret(ctx context.Context,
	kubeClient kubernetes.Interface,
	secret *v1.Secret) error {

	if _, err := GetSecret(ctx, kubeClient, secret.Namespace, secret.Name); err != nil {
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

// CompileRegistryAuthSecret creates a secret object with docker config json
func CompileRegistryAuthSecret(token *registry.Token) (*v1.Secret, error) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      token.SecretName,
			Namespace: token.Namespace,
		},
		Type: "kubernetes.io/dockerconfigjson",
	}

	auths := map[string]RegistryAuth{}
	auths[token.RegistryUri] = RegistryAuth{
		Auth: token.Auth,
	}

	configJSON, err := json.Marshal(DockerConfigJSON{Auths: auths})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal docker config json")
	}

	secret.Data = map[string][]byte{
		".dockerconfigjson": configJSON,
	}

	return secret, nil
}
