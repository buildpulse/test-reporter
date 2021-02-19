package logger

import (
	"bytes"
	"io"
	"log"
)

// A Logger represents a mechanism for logging. 🙃
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Text() string
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

// Text returns a string concatenation of all of the log's entries.
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
