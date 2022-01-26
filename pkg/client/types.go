package client

import "github.com/aws/aws-sdk-go/service/ecr"

type Token struct {
	SecretName  string
	AccessToken *string
	Endpoints   []string
}

type ECRToken struct {
	Token     Token
	ECRClient *ecr.ECR
}
