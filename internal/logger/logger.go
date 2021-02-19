package logger

import (
	"bytes"
	"io"
	"log"
)

// Logger TODO Add docs
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Text() string // TODO Probably document what this does
}

type logger struct {
	buffer *bytes.Buffer
	log    *log.Logger
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.log.Printf(format, v...)
}

func (l *logger) Println(v ...interface{}) {
	l.log.Println(v...)
}

func (l *logger) Text() string {
	return l.buffer.String()
}

// New TODO Add docs
func New(writers ...io.Writer) Logger {
	var buffer bytes.Buffer

	var logWriters []io.Writer
	logWriters = append(logWriters, &buffer)
	logWriters = append(logWriters, writers...)
	w := io.MultiWriter(logWriters...)

	return &logger{
		buffer: &buffer,
		log:    log.New(w, "<buildpulse> ", 0), // TODO: Do I have a _current_ need for any of these args to be configurable outside of this method?
	}
}
