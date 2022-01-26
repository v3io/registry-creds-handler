package registry

const (
	ECRRegistryKind string = "ecr"
)

type Token struct {
	SecretName  string
	AccessToken *string
	Endpoints   []string
}
