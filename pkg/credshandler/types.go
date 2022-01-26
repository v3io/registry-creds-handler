package credshandler

type PostECRAuthTokenParams struct {
	Region          string   `json:"region,omitempty"`
	AssumeRole      string   `json:"assumeRole,omitempty"`
	AccessKeyID     string   `json:"accessKeyID,omitempty"`
	SecretAccessKey string   `json:"secretAccessKey,omitempty"`
	SecretName      string   `json:"secretName,omitempty"`
	Endpoints       []string `json:"endpoints,omitempty"`
}
