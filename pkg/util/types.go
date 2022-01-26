package util

type DockerConfigJSON struct {
	Auths map[string]RegistryAuth `json:"auths,omitempty"`
}

type RegistryAuth struct {
	Auth string `json:"auth"`
}
