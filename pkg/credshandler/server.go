package credshandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/v3io/registry-creds-handler/pkg/client"
	"github.com/v3io/registry-creds-handler/pkg/util"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	"k8s.io/client-go/kubernetes"
)

type Server struct {
	logger        logger.Logger
	listenAddress string
	kubeClientSet *kubernetes.Clientset
	ecrClient     *client.ECRClient
}

func NewServer(logger logger.Logger,
	listenAddress string,
	kubeClientSet *kubernetes.Clientset,
	ecrClient *client.ECRClient) (*Server, error) {

	return &Server{
		logger:        logger.GetChild("server"),
		listenAddress: listenAddress,
		ecrClient:     ecrClient,
		kubeClientSet: kubeClientSet,
	}, nil
}

func (s *Server) Start() error {
	s.logger.Info("Server starting...")
	// TODO: goroutine to handle refreshing

	// register endpoints
	http.HandleFunc("/ecr/token", s.handleECRToken)

	if err := http.ListenAndServe(s.listenAddress, nil); err != nil {
		return errors.Wrap(err, "Failed while listening to incoming requests")
	}
	return nil
}

func (s *Server) handleECRToken(writer http.ResponseWriter, req *http.Request) {
	s.logger.InfoWith("Received ecr token request", "request", req)

	switch req.Method {
	case http.MethodPost:
		s.handlePostECRToken(writer, req)
	case http.MethodDelete:
		// TODO: Remove the secret.
	default:
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePostECRToken(writer http.ResponseWriter, req *http.Request) {
	s.logger.Info("Handling POST ECR token request")

	body, err := req.GetBody()
	if err != nil {
		s.logger.WarnWith("Failed to get request body", "error", err.Error())
		http.Error(writer, "Failed to get request body", http.StatusBadRequest)
		return
	}

	params := PostECRAuthTokenParams{}
	err = json.NewDecoder(body).Decode(&params)
	if err != nil {
		s.logger.WarnWith("Failed to decode json request body", "error", err.Error())
		http.Error(writer, "Received unexpected request body", http.StatusBadRequest)
		return
	}

	err = s.validatePostECRTokenParams(params)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: get namespace
	secret, err := util.GetSecret(req.Context(), s.kubeClientSet, "", params.SecretName)
	if err == nil && secret != nil {
		http.Error(writer, fmt.Sprintf("Secret `%s` already exists", params.SecretName), http.StatusConflict)
		return
	}

	awsEcrClient := s.ecrClient.CreateECRClient(params.Region, params.AssumeRole, params.AccessKeyID, params.SecretAccessKey)

	accessToken, err := s.ecrClient.GetAuthorizationToken(awsEcrClient)
	if err != nil {
		http.Error(writer, "Failed to get ECR authorization token", http.StatusBadRequest)
		return
	}

	token := client.Token{
		SecretName:  params.SecretName,
		AccessToken: accessToken,
		Endpoints:   params.Endpoints,
	}
	secretObj := util.GenerateSecretObj(token)
	err = util.CreateSecret(req.Context(), s.kubeClientSet, secretObj)
	if err != nil {
		s.logger.WarnWith("Failed to create secret", "error", err.Error())
		http.Error(writer, "Failed to create secret", http.StatusInternalServerError)
		return
	}
}

func (s *Server) validatePostECRTokenParams(params PostECRAuthTokenParams) error {
	if params.Region == "" {
		return errors.New("AWS Region is required")
	}
	if params.AccessKeyID == "" {
		return errors.New("AWS Access Key ID is required")
	}
	if params.SecretAccessKey == "" {
		return errors.New("AWS Secret Access Key is required")
	}
	if params.SecretName == "" {
		return errors.New("Token Secret Name is required")
	}

	return nil
}
