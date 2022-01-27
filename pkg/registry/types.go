package registry

const (
	ECRRegistryKind string = "ecr"
)

type Token struct {
	SecretName  string
	AccessToken *string
	RegistryUri string
}

type AWSCreds struct {
	Region          string `json:"region"`
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	AssumeRole      string `json:"assumeRole,omitempty"`
}
