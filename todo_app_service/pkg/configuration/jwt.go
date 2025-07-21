package configuration

type jwtConfig struct {
	Secret []byte `envconfig:"JWT_KEY"`
}
