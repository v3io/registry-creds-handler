package main

import (
	"flag"
	"os"

	"github.com/v3io/registry-creds-handler/pkg/client"
	"github.com/v3io/registry-creds-handler/pkg/credshandler"
	"github.com/v3io/registry-creds-handler/pkg/util"

	"github.com/nuclio/errors"
	"github.com/nuclio/loggerus"
	"github.com/sirupsen/logrus"
)

func run() error {

	// args
	logLevel := flag.String("log-level", "debug", "Set handler's log level")
	listenAddress := flag.String("listen-addr", os.Getenv("LISTEN_ADDRESS"), "Port to listen on")
	kubeConfigPath := flag.String("kube-config-path", "", "Kubernetes cluster config path, If not specified uses in cluster config")
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

	ecrClient, err := client.NewClient(logger)
	if err != nil {
		return errors.Wrap(err, "Failed to create ECR client")
	}

	// start server
	server, err := credshandler.NewServer(logger, *listenAddress, kubeClientSet, ecrClient)
	if err != nil {
		return errors.Wrap(err, "Failed to create new server")
	}
	if err = server.Start(); err != nil {
		return errors.Wrap(err, "Failed to start server")
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
