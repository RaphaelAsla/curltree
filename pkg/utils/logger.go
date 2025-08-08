package utils

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"curltree/internal/config"
)

type Logger struct {
	*slog.Logger
}

func NewLogger(cfg *config.LoggingConfig) (*Logger, error) {
	level := parseLogLevel(cfg.Level)
	
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	case "file":
		file, err := os.OpenFile(cfg.OutputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		output = file
	default:
		output = os.Stdout
	}
	
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}
	
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text":
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}
	
	logger := slog.New(handler)
	
	return &Logger{logger}, nil
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (l *Logger) WithContext(component string) *Logger {
	return &Logger{l.With("component", component)}
}

func (l *Logger) WithUser(userID string) *Logger {
	return &Logger{l.With("user_id", userID)}
}

func (l *Logger) WithRequest(method, path, userAgent string) *Logger {
	return &Logger{l.With(
		"method", method,
		"path", path,
		"user_agent", userAgent,
	)}
}

func (l *Logger) LogError(err error, message string, args ...interface{}) {
	l.Error(message, append([]interface{}{"error", err}, args...)...)
}

func (l *Logger) LogRequest(method, path string, statusCode int, duration string) {
	l.Info("HTTP request",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration", duration,
	)
}

func (l *Logger) LogSSHConnection(user, remoteAddr string) {
	l.Info("SSH connection",
		"user", user,
		"remote_addr", remoteAddr,
	)
}

func (l *Logger) LogProfileAction(action, userID, username string) {
	l.Info("Profile action",
		"action", action,
		"user_id", userID,
		"username", username,
	)
}

func (l *Logger) LogRateLimit(ip string, requestsPerMinute int) {
	l.Warn("Rate limit exceeded",
		"ip", ip,
		"limit", requestsPerMinute,
	)
}