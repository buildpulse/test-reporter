package metadata

// A Logger represents a mechanism for logging. 🙃
type Logger interface {
	Printf(format string, v ...interface{})
}
