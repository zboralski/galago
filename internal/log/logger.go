// Package log provides structured logging for galago using zap.
package log

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with galago-specific helpers.
type Logger struct {
	*zap.Logger
	onTrace func(pc uint64, category, name, detail string) // trace callback for events
}

var (
	// L is the global logger instance.
	L    *Logger
	once sync.Once
)

// Init initializes the global logger with the given configuration.
// Safe to call multiple times; only the first call takes effect.
func Init(debug bool) {
	once.Do(func() {
		L = New(debug)
	})
}

// New creates a new Logger instance.
func New(debug bool) *Logger {
	var cfg zap.Config
	if debug {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	}

	// Shorter timestamps in development
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		// Fallback to no-op if config fails
		logger = zap.NewNop()
	}

	return &Logger{Logger: logger}
}

// NewNop creates a no-op logger for testing.
func NewNop() *Logger {
	return &Logger{Logger: zap.NewNop()}
}

// SetOnTrace sets the trace callback for stub events.
func (l *Logger) SetOnTrace(fn func(pc uint64, category, name, detail string)) {
	l.onTrace = fn
}

// Trace logs a stub event and calls the trace callback if set.
// This is the primary method for stubs to report their activity.
func (l *Logger) Trace(pc uint64, category, name, detail string) {
	// Always call trace callback (for trace event collection)
	if l.onTrace != nil {
		l.onTrace(pc, category, name, detail)
	}

	// Log at debug level with structured fields
	l.Debug("stub",
		zap.String("cat", category),
		zap.String("fn", name),
		zap.String("detail", detail),
		zap.Uint64("pc", pc),
	)
}

// TraceSimple logs a stub event without PC (uses 0).
func (l *Logger) TraceSimple(category, name, detail string) {
	l.Trace(0, category, name, detail)
}

// Stub logs stub installation/registration events.
func (l *Logger) Stub(msg string, fields ...zap.Field) {
	l.Debug(msg, fields...)
}

// StubInstall logs when a stub is installed at an address.
func (l *Logger) StubInstall(category, name string, addr uint64, source string) {
	l.Debug("installed",
		zap.String("cat", category),
		zap.String("fn", name),
		zap.Uint64("addr", addr),
		zap.String("src", source),
	)
}

// StubFallback logs when a fallback stub is triggered.
func (l *Logger) StubFallback(name string) {
	l.Debug("fallback",
		zap.String("fn", name),
		zap.String("ret", "0"),
	)
}

// DetectorActivate logs when a detector is activated.
func (l *Logger) DetectorActivate(name, description string) {
	l.Info("detector",
		zap.String("name", name),
		zap.String("desc", description),
	)
}

// DetectorRegister logs when a detector is registered.
func (l *Logger) DetectorRegister(name, description string, patterns []string) {
	l.Debug("detector registered",
		zap.String("name", name),
		zap.String("desc", description),
		zap.Strings("patterns", patterns),
	)
}

// WithCategory returns a logger with the category field preset.
func (l *Logger) WithCategory(category string) *Logger {
	return &Logger{
		Logger:  l.Logger.With(zap.String("cat", category)),
		onTrace: l.onTrace,
	}
}

// Hex formats a uint64 as hex string for logging.
func Hex(addr uint64) string {
	return "0x" + hexString(addr)
}

func hexString(v uint64) string {
	const digits = "0123456789abcdef"
	if v == 0 {
		return "0"
	}
	buf := make([]byte, 16)
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = digits[v&0xf]
		v >>= 4
	}
	return string(buf[i:])
}

// Field helpers for common patterns.

// Addr creates an address field.
func Addr(addr uint64) zap.Field {
	return zap.String("addr", Hex(addr))
}

// Size creates a size field.
func Size(size uint64) zap.Field {
	return zap.Uint64("size", size)
}

// Ptr creates a pointer field.
func Ptr(name string, ptr uint64) zap.Field {
	return zap.String(name, Hex(ptr))
}

// Fn creates a function name field.
func Fn(name string) zap.Field {
	return zap.String("fn", name)
}
