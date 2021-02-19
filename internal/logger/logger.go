package logger

import (
	"log"
	"os"
)

// Logger TODO Add docs
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type logger struct {
	log *log.Logger
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.log.Printf(format, v...)
}

func (l *logger) Println(v ...interface{}) {
	l.log.Println(v...)
}

// New TODO Add docs
func New() Logger {
	return &logger{
		log: log.New(os.Stdout, "<buildpulse> ", 0), // TODO: Do I have a _current_ need for any of these args to be configurable outside of this method?
	}
}
