package logging

import (
	"context"
	"github.com/google/uuid"
	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"os"
)

// Logger context key to define a new type so ctx strings from other packages cant conflict with it
type ctxKey string

const loggerKey = ctxKey("logger")

type Logger struct {
	*zap.SugaredLogger
	LevelId string
}

func (logger *Logger) With(fields ...interface{}) *Logger {
	return &Logger{
		logger.SugaredLogger.With(fields...),
		logger.LevelId,
	}
}

func (logger *Logger) WithRequestId(id string) *Logger {
	return logger.With("request-id", id)
}

func (logger *Logger) WithRequestEndpointName(name string) *Logger {
	return logger.With("endpoint", name)
}

func (logger *Logger) WithStringKVP(key, value string) *Logger {
	return &Logger{
		logger.SugaredLogger.With(zap.String(key, value)),
		logger.LevelId,
	}
}

func (logger *Logger) WithInt(key string, value int) *Logger {
	return &Logger{
		logger.SugaredLogger.With(zap.Int(key, value)),
		logger.LevelId,
	}
}

func LoggerFactoryFor(component string) *Logger {
	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, os.Stdout, zap.InfoLevel)
	newId := uuid.New().String()
	logger := zap.New(core, zap.AddCaller()).
		With(
			zap.String("component", component),
			zap.String("id", newId),
		).Sugar()
	return &Logger{logger, newId}
}

func SetLogger(parent context.Context, logger *Logger) context.Context {
	return context.WithValue(parent, loggerKey, logger)
}

// GetLogger gets the logger from context. If the context has no logger, it returns a new dev logger
func GetLogger(ctx context.Context) *Logger {
	v := ctx.Value(loggerKey)
	if logger, ok := v.(*Logger); ok {
		return logger
	}
	//Default is a basic logger
	devModeKeyValue := zap.String("mode", "dev")
	devLogger, err := zap.NewDevelopment()
	if err != nil {
		panic("failed to initialize dev logger")
	}
	return &Logger{
		SugaredLogger: devLogger.With(devModeKeyValue).Sugar(),
		LevelId:       "dev"}
}
