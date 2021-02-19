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

// New returns a Logger that writes to writers and an in-memory store for
// on-demand access to log entries via the Text() method.
func New(writers ...io.Writer) Logger {
	var buffer bytes.Buffer

	var logWriters []io.Writer
	logWriters = append(logWriters, &buffer)
	logWriters = append(logWriters, writers...)
	w := io.MultiWriter(logWriters...)

	return &logger{
		buffer: &buffer,
		log:    log.New(w, "<buildpulse> ", 0),
	}
}
