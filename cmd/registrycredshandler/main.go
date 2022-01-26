package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/v3io/registry-creds-handler/pkg"
	"github.com/v3io/registry-creds-handler/pkg/registry"
	"github.com/v3io/registry-creds-handler/pkg/util"

	"github.com/nuclio/errors"
	"github.com/nuclio/loggerus"
	"github.com/sirupsen/logrus"
	"github.com/v3io/version-go"
)

func run() error {
	fmt.Printf("Handler version:\n%#v", version.Get().String())

	// args
	logLevel := flag.String("log-level", "debug", "Set handler's log level")
	registryKind := flag.String("registry-kind", "ecr", "Docker registry kind (ecr)")
	secretName := flag.String("secret-name", "", "Secret name must be unique")
	namespace := flag.String("namespace", "", "Kubernetes namespace of secret")
	registryURIs := flag.String("registry-uris", "", "Comma seperated list of registry URIs")
	refreshRate := flag.Int64("refresh-rate", 60, "Secret refresh rate in min, default is 60 min")
	kubeConfigPath := flag.String("kube-config-path", "", "Kubernetes cluster config path, If not specified uses in cluster config")
	creds := flag.String("creds", "", "Credentials to retrieve registry authorization token in JSON format, entries must be in lowerCamelCase")

	flag.Parse()

	// logger conf
	parsedLogLevel, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		return errors.Wrap(err, "Failed to parse log level")
	}

	logger, err := loggerus.NewJSONLoggerus("main", parsedLogLevel, os.Stdout)
	if err != nil {
		return errors.Wrap(err, "Failed to create new logger")
	}

	// create clients
	kubeClientSet, err := util.NewKubeClientSet(*kubeConfigPath)
	if err != nil {
		return errors.Wrap(err, "Failed to create k8s clientset")
	}

	ecr, err := registry.NewECRRegistry(logger,
		*secretName,
		*namespace,
		*creds,
		strings.Split(*registryURIs, ","))
	if err != nil {
		return errors.Wrap(err, "Failed to create ECR")
	}

	// validate the requested registry kind params
	switch *registryKind {
	case registry.ECRRegistryKind:
		err := ecr.EnrichAndValidateECRParams()
		if err != nil {
			return errors.Wrap(err, "Failed ECR params validation")
		}
	default:
		return errors.New("Received unsupported registry kind")
	}

	// start handler
	handler, err := pkg.NewHandler(logger, kubeClientSet, ecr, *refreshRate, *registryKind)
	if err != nil {
		return errors.Wrap(err, "Failed to create new handler")
	}
	if err = handler.Run(); err != nil {
		return errors.Wrap(err, "Failed to start handler")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		errors.PrintErrorStack(os.Stderr, err, 5)
		os.Exit(1)
	}

	os.Exit(0)
}
