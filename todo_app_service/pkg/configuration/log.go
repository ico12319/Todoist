package configuration

import (
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

type logConfig struct {
	Level  string `envconfig:"APP_LOG_LEVEL" default:"debug"`
	Format string `envconfig:"APP_LOG_FORMAT" default:"text"`
}

type logKey struct{}

var (
	mu = sync.Mutex{}
	C  = LoggerFromContext
)

func SetUpLogger(ctx context.Context, cfg logConfig) (context.Context, error) {
	mu.Lock()
	defer mu.Unlock()

	logEntry := logrus.NewEntry(logrus.StandardLogger())
	level, err := logrus.ParseLevel(cfg.Level)

	if err != nil {
		return nil, err
	}

	logEntry.Logger.SetLevel(level)
	logEntry.Logger.SetOutput(os.Stdout)

	if cfg.Format == constants.JSON_FORMAT {
		logEntry.Logger.SetFormatter(&logrus.JSONFormatter{
			PrettyPrint: true,
		})
	} else if cfg.Format == constants.TEXT_FORMAT {
		logEntry.Logger.SetFormatter(&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
	}

	return ContextWithLogger(ctx, logEntry), nil
}

func ContextWithLogger(ctx context.Context, logEntry *logrus.Entry) context.Context {
	return context.WithValue(ctx, logKey{}, logEntry)
}

func LoggerFromContext(ctx context.Context) *logrus.Entry {
	mu.Lock()
	defer mu.Unlock()

	logEntry, ok := ctx.Value(logKey{}).(*logrus.Entry)
	if !ok {
		logEntry = logrus.NewEntry(logrus.StandardLogger())
	}
	return logEntry
}
