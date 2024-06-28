package log

import "log/slog"

// Format for log output
type LogFormat string

const (
	// Use logfmt
	LogFormat_Logfmt LogFormat = "logfmt"

	// Use JSON
	LogFormat_Json LogFormat = "json"
)

// Options for logging
type LoggerOptions struct {
	// === Lumberjack Options ===

	// The maximum size (in megabytes) of the log file before it gets rotated
	MaxSize int

	// The maximum number of old log files to retain.
	// Use 0 to retain all backups.
	MaxBackups int

	// The maximum number of days to retain old log files based on the timestamp encoded in their filename.
	// Use 0 to always preserve old logs.
	MaxAge int

	// True to format the timestamps in backup files in the computer's local time; false to format in UTC
	LocalTime bool

	// True to compress rotated log files using gzip
	Compress bool

	// === Slog Options ===

	// The format to use when printing logs
	Format LogFormat

	// The minimum record level that will be logged
	Level slog.Level

	// True to include the source code position of the log statement in log messages
	AddSource bool

	// True to enable HTTP request tracing
	EnableHttpTracing bool
}
