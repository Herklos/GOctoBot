package logging

import (
    "fmt"
    "log"
    "sync"

    octobot_commons "octobot_commons/octobot_commons"
)

const (
    LOG_DATABASE           = "log_db"
    LOG_NEW_ERRORS_COUNT   = "log_new_errors_count"
    BACKTESTING_NEW_ERRORS_COUNT = "log_backtesting_errors_count"
)

const (
    LevelDebug   = 10
    LevelInfo    = 20
    LevelWarning = 30
    LevelError   = 40
    LevelCritical = 50
    LevelFatal   = 50
)

type logEntry struct {
    Time    string
    Level   string
    Source  string
    Message string
}

var (
    logsDatabase = map[string]any{
        LOG_DATABASE:           []logEntry{},
        LOG_NEW_ERRORS_COUNT:   0,
        BACKTESTING_NEW_ERRORS_COUNT: 0,
    }

    errorNotifierCallbacks []func()

    logsMaxCount = 1000
    storedLogMinLevel = LevelWarning
    enableWebInterfaceLogs = true
    errorPublicationEnabled = true
    shouldPublishLogsWhenReEnabled = false

    errorCallback = defaultCallback
    logCallback func(string) string

    globalLoggerLevel = LevelInfo
    mu sync.Mutex
)

func defaultCallback(_ error, _ string) {}

func SetGlobalLoggerLevel(level int, handlerLevels []int) {
    mu.Lock()
    defer mu.Unlock()
    globalLoggerLevel = level
}

func GetGlobalLoggerLevel() int {
    mu.Lock()
    defer mu.Unlock()
    return globalLoggerLevel
}

func TemporaryLogLevel(level int, fn func()) {
    previous := GetGlobalLoggerLevel()
    SetGlobalLoggerLevel(level, nil)
    defer SetGlobalLoggerLevel(previous, nil)
    fn()
}

func GetLoggerLevelPerHandler() []int {
    return []int{GetGlobalLoggerLevel()}
}

func GetLogger(loggerName string) *BotLogger {
    return &BotLogger{LoggerName: loggerName}
}

func SetLoggingLevel(loggerNames []string, level int) {
    _ = loggerNames
    SetGlobalLoggerLevel(level, nil)
}

func AddLog(level int, source any, message string, keepLog bool, callNotifiers bool) {
    mu.Lock()
    defer mu.Unlock()

    if keepLog {
        entries := logsDatabase[LOG_DATABASE].([]logEntry)
        entries = append(entries, logEntry{
            Time:    octobot_commons.GetNowTime("2006-01-02 15:04:05", true),
            Level:   levelName(level),
            Source:  fmt.Sprint(source),
            Message: message,
        })
        if len(entries) > logsMaxCount {
            entries = entries[1:]
        }
        logsDatabase[LOG_DATABASE] = entries
        if level >= LevelError {
            logsDatabase[LOG_NEW_ERRORS_COUNT] = logsDatabase[LOG_NEW_ERRORS_COUNT].(int) + 1
            logsDatabase[BACKTESTING_NEW_ERRORS_COUNT] = logsDatabase[BACKTESTING_NEW_ERRORS_COUNT].(int) + 1
        }
    }

    if callNotifiers {
        for _, cb := range errorNotifierCallbacks {
            cb()
        }
    }
}

func GetErrorsCount(counter string) int {
    mu.Lock()
    defer mu.Unlock()
    return logsDatabase[counter].(int)
}

func ResetErrorsCount(counter string) {
    mu.Lock()
    defer mu.Unlock()
    logsDatabase[counter] = 0
}

func RegisterErrorNotifier(callback func()) {
    errorNotifierCallbacks = append(errorNotifierCallbacks, callback)
}

type BotLogger struct {
    LoggerName string
}

func (l *BotLogger) Debug(message string, args ...any) {
    message = l.processLogCallback(message)
    log.Printf("[DEBUG] "+message, args...)
    l.publishLogIfNecessary(message, LevelDebug)
}

func (l *BotLogger) Info(message string, args ...any) {
    message = l.processLogCallback(message)
    log.Printf("[INFO] "+message, args...)
    l.publishLogIfNecessary(message, LevelInfo)
}

func (l *BotLogger) Warning(message string, args ...any) {
    message = l.processLogCallback(message)
    log.Printf("[WARNING] "+message, args...)
    l.publishLogIfNecessary(message, LevelWarning)
}

func (l *BotLogger) Error(message string, args ...any) {
    message = l.processLogCallback(message)
    log.Printf("[ERROR] "+message, args...)
    l.publishLogIfNecessary(message, LevelError)
    l.postCallbackIfNecessary(nil, message, false)
}

func (l *BotLogger) ErrorWithSkip(message string, skipPostCallback bool, args ...any) {
    message = l.processLogCallback(message)
    log.Printf("[ERROR] "+message, args...)
    l.publishLogIfNecessary(message, LevelError)
    l.postCallbackIfNecessary(nil, message, skipPostCallback)
}

func (l *BotLogger) Exception(exception error, publishErrorIfNecessary bool, errorMessage string, includeExceptionName bool, skipPostCallback bool) {
    originErrorMessage := errorMessage
    if errorMessage != "" {
        errorMessage = l.processLogCallback(errorMessage)
    }
    _ = originErrorMessage
    octobot_commons.SummarizeExceptionHTMLCauseIfRelevant(exception, 0)
    log.Printf("[EXCEPTION] %v", exception)
    if publishErrorIfNecessary {
        message := originErrorMessage
        if message == "" {
            message = fmt.Sprint(exception)
        } else if includeExceptionName {
            message = fmt.Sprintf("%s (error: %T)", message, exception)
        }
        l.ErrorWithSkip(message, true)
        _ = octobot_commons.EXCEPTION_DESC
        _ = octobot_commons.IS_EXCEPTION_DESC
    }
    l.postCallbackIfNecessary(exception, errorMessage, skipPostCallback)
}

func (l *BotLogger) Critical(message string, args ...any) {
    message = l.processLogCallback(message)
    log.Printf("[CRITICAL] "+message, args...)
    l.publishLogIfNecessary(message, LevelCritical)
}

func (l *BotLogger) Fatal(message string, args ...any) {
    message = l.processLogCallback(message)
    log.Fatalf("[FATAL] "+message, args...)
    l.publishLogIfNecessary(message, LevelFatal)
}

func (l *BotLogger) Disable(disabled bool) {
    _ = disabled
}

func (l *BotLogger) processLogCallback(message string) string {
    if logCallback == nil {
        return message
    }
    return logCallback(message)
}

func (l *BotLogger) publishLogIfNecessary(message string, level int) {
    if enableWebInterfaceLogs && storedLogMinLevel <= level && GetGlobalLoggerLevel() <= level {
        l.webInterfacePublishLog(message, level)
        if !errorPublicationEnabled && level >= LevelError {
            shouldPublishLogsWhenReEnabled = true
        }
    }
}

func (l *BotLogger) webInterfacePublishLog(message string, level int) {
    AddLog(level, l.LoggerName, message, true, errorPublicationEnabled)
}

func (l *BotLogger) postCallbackIfNecessary(exception error, errorMessage string, skipPostCallback bool) {
    if !skipPostCallback {
        errorCallback(exception, errorMessage)
    }
}

func RegisterErrorCallback(callback func(error, string)) {
    errorCallback = callback
}

func DefaultCallback() func(error, string) {
    return defaultCallback
}

func ErrorCallback() func(error, string) {
    return errorCallback
}

func RegisterLogCallback(callback func(string) string) {
    logCallback = callback
}

func GetBacktestingErrorsCount() int {
    return GetErrorsCount(BACKTESTING_NEW_ERRORS_COUNT)
}

func ResetBacktestingErrors() {
    ResetErrorsCount(BACKTESTING_NEW_ERRORS_COUNT)
}

func SetErrorPublicationEnabled(enabled bool) {
    errorPublicationEnabled = enabled
    if enabled && shouldPublishLogsWhenReEnabled {
        AddLog(LevelError, nil, "", false, true)
    } else {
        shouldPublishLogsWhenReEnabled = false
    }
}

func SetEnableWebInterfaceLogs(enabled bool) {
    enableWebInterfaceLogs = enabled
}

func levelName(level int) string {
    if level == LevelFatal {
        return "CRITICAL"
    }
    switch level {
    case LevelDebug:
        return "DEBUG"
    case LevelInfo:
        return "INFO"
    case LevelWarning:
        return "WARNING"
    case LevelError:
        return "ERROR"
    case LevelCritical:
        return "CRITICAL"
    default:
        return fmt.Sprintf("LEVEL_%d", level)
    }
}
