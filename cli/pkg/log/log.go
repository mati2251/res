package log

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Level string

const (
	LevelStdout Level = "stdout" // the script output
	LevelStderr Level = "stderr" // the script error output
	LevelCode   Level = "code"   // the script exit code
	LevelInfo   Level = "info"   // general information
	LevelError  Level = "error"  // error information
	Green             = "\033[32m"
	Red               = "\033[31m"
	Reset             = "\033[0m"
)

type Logger interface {
	Printf(level Level, format string, args ...any)
	Writer(Level) io.Writer
}

type HumanLogger struct {
	Out io.Writer
	Err io.Writer
}

type HumanLoggerWriter struct {
	humanLogger HumanLogger
	level       Level
}

func (h HumanLoggerWriter) Write(p []byte) (n int, err error) {
	h.humanLogger.Printf(h.level, "%s", string(p))
	return len(p), nil
}

func (l HumanLogger) Printf(level Level, format string, args ...any) {
	switch level {
	case LevelInfo:
		_, _ = l.Err.Write(fmt.Appendf(nil, Green+format+Reset, args...))
	case LevelError:
		_, _ = l.Err.Write(fmt.Appendf(nil, Red+format+Reset, args...))
	default:
		_, _ = l.Out.Write(fmt.Appendf(nil, format, args...))
	}
}

func (l HumanLogger) Writer(level Level) io.Writer {
	return HumanLoggerWriter{
		humanLogger: l,
		level:       level,
	}
}

type JsonLine struct {
	Level     Level  `json:"level"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

type JsonLogger struct {
	Out io.Writer
}

type JsonLoggerWriter struct {
	jsonLogger JsonLogger
	level      Level
}

func (l JsonLogger) Printf(level Level, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	jsonLine := JsonLine{
		Level:     Level(level),
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   msg,
	}
	jsonData, err := json.Marshal(jsonLine)
	if err != nil {
		_, _ = fmt.Fprintf(l.Out, "Error marshalling JSON: %v\n", err)
		return
	}
	_, _ = l.Out.Write(fmt.Appendf(nil, "%s\n", jsonData))
}

func (l JsonLoggerWriter) Write(p []byte) (n int, err error) {
	l.jsonLogger.Printf(l.level, "%s", string(p))
	return len(p), nil
}

func (l JsonLogger) Writer(level Level) io.Writer {
	return JsonLoggerWriter{
		jsonLogger: l,
		level:      level,
	}
}
