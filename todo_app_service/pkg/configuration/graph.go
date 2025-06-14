package log

type graphQlServerConfig struct {
	Port string `envconfig:"GRAPHQL_SERVER_PORT"`
}
