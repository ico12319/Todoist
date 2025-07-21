package configuration

type restConfig struct {
	Port       string `envconfig:"REST_PORT" default:"3434"`
	TodoApiUrl string `envconfig:"TODO_REST_API_URL"`
}
