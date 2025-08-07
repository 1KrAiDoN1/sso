package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

type Logger struct {
	logger  zerolog.Logger
	service string
}

func New(serviceName string) *Logger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		FormatLevel: func(i interface{}) string {
			return colorizeLevel(fmt.Sprintf("%-5s", i))
		},
		FormatMessage: func(i interface{}) string {
			return color.CyanString("message=%s", i)
		},
		FormatFieldName: func(i interface{}) string {
			return color.BlueString("%s=", i)
		},
		FormatFieldValue: func(i interface{}) string {
			return color.WhiteString("%v", i)
		},
	}
	log := zerolog.New(output).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	return &Logger{
		logger:  log,
		service: serviceName,
	}
}

func colorizeLevel(level interface{}) string {
	l := fmt.Sprintf("%-5s", level)
	switch strings.ToLower(strings.TrimSpace(l)) {
	case "debug":
		return color.BlueString(l)
	case "info":
		return color.GreenString(l)
	case "warn":
		return color.YellowString(l)
	case "error", "fatal":
		return color.RedString(l)
	default:
		return l
	}
}

func (l *Logger) Debug(func_name string, msg string) {
	l.logger.Debug().Str("func", func_name).Msg(msg)
}

func (l *Logger) Info(func_name string, msg string) {
	l.logger.Info().Str("func", func_name).Msg(msg)
}

func (l *Logger) Warn(func_name string, msg string) {
	l.logger.Warn().Str("func", func_name).Msg(msg)
}

func (l *Logger) Error(func_name string, msg string) {
	l.logger.Error().Str("func", func_name).Msg(msg)
}

func (l *Logger) Fatal(func_name string, msg string) {
	defer os.Exit(1)
	l.logger.Fatal().Str("func", func_name).Msg(msg)
}
