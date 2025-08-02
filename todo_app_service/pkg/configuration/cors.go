package configuration

type corsConfig struct {
	FrontendUrl string `envconfig:"FRONTEND_URL"`
}
