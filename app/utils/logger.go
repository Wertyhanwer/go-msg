package utils

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[API] ", log.LstdFlags),
	}
}

func (l *Logger) Info(message string) {
	l.Printf("INFO: %s", message)
}

func (l *Logger) Error(message string) {
	l.Printf("ERROR: %s", message)
}

func (l *Logger) Fatal(message string) {
	l.Printf("FATAL: %s", message)
	os.Exit(1)
}
