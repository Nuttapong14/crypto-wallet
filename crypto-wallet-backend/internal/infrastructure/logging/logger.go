package logging

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Config describes logger initialisation parameters.
type Config struct {
	Level     string
	Format    string
	AddSource bool
	Output    io.Writer
}

// NewLogger constructs a slog.Logger based on the provided configuration.
func NewLogger(cfg Config) (*slog.Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	options := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     level,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			// Ensure consistent timestamp formatting.
			if attr.Key == slog.TimeKey {
				if t, ok := attr.Value.Any().(time.Time); ok {
					attr.Value = slog.StringValue(t.UTC().Format(time.RFC3339Nano))
				}
			}
			return attr
		},
	}

	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(cfg.Format)) {
	case "", "json":
		handler = slog.NewJSONHandler(cfg.Output, options)
	case "text":
		handler = slog.NewTextHandler(cfg.Output, options)
	default:
		return nil, errors.New("logging: unsupported format (expected json or text)")
	}

	return slog.New(handler), nil
}

// WithComponent attaches a component attribute to the supplied logger.
func WithComponent(logger *slog.Logger, component string) *slog.Logger {
	if logger == nil {
		return slog.Default()
	}
	component = strings.TrimSpace(component)
	if component == "" {
		return logger
	}
	return logger.With(slog.String("component", component))
}

func parseLevel(value string) (slog.Leveler, error) {
	if strings.TrimSpace(value) == "" {
		return slog.LevelInfo, nil
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return nil, errors.New("logging: invalid level (debug, info, warn, error)")
	}
}
