package log

type jwtConfig struct {
	Secret []byte `envconfig:"JWT_SECRET"`
}
