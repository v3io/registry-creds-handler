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
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/v3io/registry-creds-handler/pkg/common"
	"github.com/v3io/registry-creds-handler/pkg/registry/factory"
	"github.com/v3io/registry-creds-handler/pkg/registrycredshandler"

	"github.com/nuclio/errors"
	"github.com/v3io/version-go"
)

func run() error {

	// args
	verbose := flag.Bool("verbose", false, "Allow verbosity logging")
	registryKind := flag.String("registry-kind", "ecr", "Docker registry kind to authenticate against (Default: ecr)")
	secretName := flag.String("secret-name", "", "Secret name to create or update with refreshed registry credentials")
	namespace := flag.String("namespace", "", "Kubernetes namespace to create secret on")
	registryUri := flag.String("registry-uri", "", "Registry URI to use for authentication")
	refreshRate := flag.Int64("refresh-rate", 60, "Refresh credentials rate in min (Default: 60 minutes)")
	kubeConfigPath := flag.String("kubeconfig-path", "", "Kubernetes config path, If not specified uses in cluster config")
	creds := flag.String("creds", "", "Credentials to retrieve registry authorization token in JSON format, entries must be in lowerCamelCase")
	showVersion := flag.Bool("version", false, "Show version in j and exit")
	logsFormat := flag.String("logs-format", "humanreadable", "Logging format (json|humanreadable) (Default: humanreadable)")

	flag.Parse()

	if *showVersion {
		encodedVersionInfo, _ := json.Marshal(version.Get())
		fmt.Print(string(encodedVersionInfo))
		return nil
	}

	logger, err := common.CreateLogger("main", *verbose, os.Stdout, *logsFormat)
	if err != nil {
		return errors.Wrap(err, "Failed to create logger")
	}

	// create clients
	kubeClientSet, err := common.NewKubeClientSet(*kubeConfigPath)
	if err != nil {
		return errors.Wrap(err, "Failed to create k8s clientset")
	}

	// create registry
	registry, err := factory.CreateRegistry(logger, *registryKind, *secretName, *namespace, *creds, *registryUri)
	if err != nil {
		return errors.Wrap(err, "Failed to create k8s clientset")
	}

	// start handler
	handler, err := registrycredshandler.NewHandler(logger, kubeClientSet, registry, *refreshRate, *registryKind)
	if err != nil {
		return errors.Wrap(err, "Failed to create new handler")
	}
	if err = handler.Start(); err != nil {
		return errors.Wrap(err, "Failed to start handler")
	}

	select {}
}

func main() {
	if err := run(); err != nil {
		errors.PrintErrorStack(os.Stderr, err, 5)
		os.Exit(1)
	}

	os.Exit(0)
}
