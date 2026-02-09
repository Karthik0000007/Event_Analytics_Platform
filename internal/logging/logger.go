package logging

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
	Service string
}

func New(service string) *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", 0),
		Service: service,
	}
}

func (l *Logger) Info(msg string, fields map[string]any) {
	l.log("INFO", msg, fields)
}

func (l *Logger) Error(msg string, fields map[string]any) {
	l.log("ERROR", msg, fields)
}

func (l *Logger) log(level, msg string, fields map[string]any) {
	fields["level"] = level
	fields["service"] = l.Service
	fields["message"] = msg
	l.Logger.Println(fields)
}
