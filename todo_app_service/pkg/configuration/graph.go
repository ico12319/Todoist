package configuration

type graphQlServerConfig struct {
	Port string `envconfig:"GRAPHQL_SERVER_PORT" default:"8090"`
}
