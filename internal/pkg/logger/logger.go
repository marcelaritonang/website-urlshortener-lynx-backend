package logger

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return f.Function, f.File
		},
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
}

type Fields map[string]interface{}

func WithContext(ctx context.Context) *logrus.Entry {
	return log.WithContext(ctx)
}

func Info(msg string, fields Fields) {
	if fields == nil {
		log.Info(msg)
		return
	}
	log.WithFields(logrus.Fields(fields)).Info(msg)
}

func Error(msg string, err error, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	fields["error"] = err.Error()
	log.WithFields(logrus.Fields(fields)).Error(msg)
}

func Debug(msg string, fields Fields) {
	if fields == nil {
		log.Debug(msg)
		return
	}
	log.WithFields(logrus.Fields(fields)).Debug(msg)
}

func Fatal(msg string, err error, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	fields["error"] = err.Error()
	log.WithFields(logrus.Fields(fields)).Fatal(msg)
}
