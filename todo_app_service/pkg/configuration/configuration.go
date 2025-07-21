package configuration

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"os"
	"sync"
)

// config is implemented with singleton design pattern so we can share state!

type Config struct {
	DbConfig       databaseConfig
	LogConfig      logConfig
	RestConfig     restConfig
	GraphConfig    graphQlServerConfig
	OauthConfig    *oauth2.Config
	JwtConfig      jwtConfig
	ActivityConfig activityConfig
}

var (
	once     sync.Once
	instance *Config
)

func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{}
	})
	return instance
}

var config *Config = GetInstance()

func init() {
	if err := godotenv.Load("/Users/I763039/Library/CloudStorage/OneDrive-SAPSE/Documents/GitHub/Todo-List/internProject/.env"); err != nil {
		panic(err)
	}

	if err := envconfig.Process("", &config.LogConfig); err != nil {
		panic(err)
	}

	ctx, err := SetUpLogger(context.Background(), config.LogConfig)
	if err != nil {
		panic(err)
	}
	_ = ctx

	if err = envconfig.Process("", &config.DbConfig); err != nil {
		panic(err)
	}

	if err = envconfig.Process("", &config.RestConfig); err != nil {
		panic(err)
	}

	if err = envconfig.Process("", &config.GraphConfig); err != nil {
		panic(err)
	}

	config.OauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"read:user", "user:email", "read:org"},
		Endpoint:     github.Endpoint,
		RedirectURL:  os.Getenv("CALLBACK_URL"),
	}

	if err = envconfig.Process("", &config.JwtConfig); err != nil {
		panic(err)
	}

	if err = envconfig.Process("", &config.ActivityConfig); err != nil {
		panic(err)
	}
}
