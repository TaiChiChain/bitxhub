package loggers

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slog"
)

var _ slog.Handler = (*LogrusHandler)(nil)

var levelMap = map[slog.Level]logrus.Level{
	slog.LevelDebug: logrus.DebugLevel,
	slog.LevelInfo:  logrus.InfoLevel,
	slog.LevelWarn:  logrus.WarnLevel,
	slog.LevelError: logrus.ErrorLevel,
}

var levelMapReverse = map[logrus.Level]slog.Level{
	logrus.DebugLevel: slog.LevelDebug,
	logrus.InfoLevel:  slog.LevelInfo,
	logrus.WarnLevel:  slog.LevelWarn,
	logrus.ErrorLevel: slog.LevelError,
}

type LogrusHandler struct {
	Logger *logrus.Entry
	Level  slog.Leveler
}

func (h *LogrusHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.Level.Level()
}

func (h *LogrusHandler) Handle(ctx context.Context, record slog.Record) error {
	level := levelMap[record.Level]

	args := make(map[string]any)
	record.Attrs(func(attr slog.Attr) bool {
		args[attr.Key] = attr.Value
		return true
	})

	h.Logger.
		WithContext(ctx).
		WithTime(record.Time).
		WithFields(args).
		Log(level, record.Message)
	return nil
}

func (h *LogrusHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LogrusHandler{
		Logger: h.Logger,
		Level:  h.Level,
	}
}

func (h *LogrusHandler) WithGroup(name string) slog.Handler {
	return &LogrusHandler{
		Logger: h.Logger,
		Level:  h.Level,
	}
}
