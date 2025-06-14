package log

type restConfig struct {
	Port       string `envconfig:"REST_PORT"`
	TodoApiUrl string `envconfig:"TODO_REST_API_URL"`
}
