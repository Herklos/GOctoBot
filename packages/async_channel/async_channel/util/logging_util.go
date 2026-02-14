package util

import (
    "fmt"

    commons_logging "octobot_commons/octobot_commons/logging"
)

// Logger defines the minimal logger interface used by async_channel.
type Logger interface {
    Debugf(format string, args ...any)
    Errorf(format string, args ...any)
}

type commonsLogger struct {
    logger *commons_logging.BotLogger
}

func (l commonsLogger) Debugf(format string, args ...any) {
    l.logger.Debug(fmt.Sprintf(format, args...))
}

func (l commonsLogger) Errorf(format string, args ...any) {
    l.logger.Error(fmt.Sprintf(format, args...))
}

// GetLogger returns a logger implementation.
func GetLogger(name string) Logger {
    return commonsLogger{logger: commons_logging.GetLogger(name)}
}
