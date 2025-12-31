package audit

import (
    "context"
    "log/slog"
    "time"
)

// Entry represents an auditable action within the platform.
type Entry struct {
    ActorID  any
    Action   string
    TargetID string
    Metadata map[string]any
    Occurred time.Time
}

// Logger writes audit entries to the configured destination (stdout by default).
type Logger struct {
    logger *slog.Logger
}

// NewLogger constructs an Audit Logger.
func NewLogger(logger *slog.Logger) *Logger {
    if logger == nil {
        logger = slog.Default()
    }
    return &Logger{logger: logger.With(slog.String("component", "audit"))}
}

// Record persists an audit entry.
func (l *Logger) Record(_ context.Context, entry Entry) error {
    if entry.Occurred.IsZero() {
        entry.Occurred = time.Now().UTC()
    }
    l.logger.Info("audit entry", slog.Any("actor", entry.ActorID), slog.String("action", entry.Action), slog.String("target", entry.TargetID), slog.Any("metadata", entry.Metadata), slog.Time("occurred", entry.Occurred))
    return nil
}
