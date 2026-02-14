package util

import "log"

// Logger defines the minimal logger interface used by async_channel.
type Logger interface {
    Debugf(format string, args ...any)
    Errorf(format string, args ...any)
}

type stdLogger struct {
    name string
}

func (l stdLogger) Debugf(format string, args ...any) {
    log.Printf("[%s] "+format, append([]any{l.name}, args...)...)
}

func (l stdLogger) Errorf(format string, args ...any) {
    log.Printf("[%s] "+format, append([]any{l.name}, args...)...)
}

// GetLogger returns a logger implementation.
func GetLogger(name string) Logger {
    return stdLogger{name: name}
}
