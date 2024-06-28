package config

import (
	"log/slog"

	"github.com/rocket-pool/node-manager-core/config/ids"
	"github.com/rocket-pool/node-manager-core/log"
)

// Configuration for the daemon loggers
type LoggerConfig struct {
	// The minimum record level that will be logged
	Level Parameter[slog.Level]

	// The format to use when printing logs
	Format Parameter[log.LogFormat]

	// True to include the source code position of the log statement in log messages
	AddSource Parameter[bool]

	// The maximum size (in megabytes) of the log file before it gets rotated
	MaxSize Parameter[uint64]

	// The maximum number of old log files to retain
	MaxBackups Parameter[uint64]

	// The maximum number of days to retain old log files based on the timestamp encoded in their filename
	MaxAge Parameter[uint64]

	// Toggle for saving rotated logs with local system time in the name vs. UTC
	LocalTime Parameter[bool]

	// Toggle for compressing rotated logs
	Compress Parameter[bool]

	// Toggle for enabling HTTP request tracing
	EnableHttpTracing Parameter[bool]
}

// Generates a new Logger configuration
func NewLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level: Parameter[slog.Level]{
			ParameterCommon: &ParameterCommon{
				ID:                ids.LoggerLevelID,
				Name:              "Log Level",
				Description:       "Select the minimum level for log messages. The lower it goes, the more verbose output the logs contain.",
				AffectsContainers: []ContainerID{ContainerID_Daemon},
			},
			Options: []*ParameterOption[slog.Level]{
				{
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Debug",
						Description: "Log debug messages - useful for development, or if something goes wrong and you need to provide extra information to supporters in order to track issues down.",
					},
					Value: slog.LevelDebug,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Info",
						Description: "Log routine info messages.",
					},
					Value: slog.LevelInfo,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Warn",
						Description: "Only log warnings or higher, skipping info messages.",
					},
					Value: slog.LevelWarn,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Error",
						Description: "Only log errors that prevent the daemon from running as expected.",
					},
					Value: slog.LevelError,
				},
			},
			Default: map[Network]slog.Level{
				Network_All: slog.LevelInfo,
			},
		},

		Format: Parameter[log.LogFormat]{
			ParameterCommon: &ParameterCommon{
				ID:                ids.LoggerFormatID,
				Name:              "Format",
				Description:       "Choose which format log messages will be printed in.",
				AffectsContainers: []ContainerID{ContainerID_Daemon},
			},
			Options: []*ParameterOption[log.LogFormat]{
				{
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Logfmt",
						Description: "Use the logfmt format, which offers a good balance of human readability and parsability. See https://www.brandur.org/logfmt for more information on this format.",
					},
					Value: log.LogFormat_Logfmt,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "JSON",
						Description: "Log messages in JSON format. Useful if you want to process your logs through other tooling.",
					},
					Value: log.LogFormat_Json,
				},
			},
			Default: map[Network]log.LogFormat{
				Network_All: log.LogFormat_Logfmt,
			},
		},

		AddSource: Parameter[bool]{
			ParameterCommon: &ParameterCommon{
				ID:                ids.LoggerAddSourceID,
				Name:              "Embed Source Location",
				Description:       "Enable this to add the source location of where the logger was called to each log message. This is mostly for development use only.",
				AffectsContainers: []ContainerID{ContainerID_Daemon},
			},
			Default: map[Network]bool{
				Network_All: false,
			},
		},

		MaxSize: Parameter[uint64]{
			ParameterCommon: &ParameterCommon{
				ID:                ids.LoggerMaxSizeID,
				Name:              "Max Log Size",
				Description:       "The max size (in megabytes) of a log file before it gets rotated out and archived.",
				AffectsContainers: []ContainerID{ContainerID_Daemon},
			},
			Default: map[Network]uint64{
				Network_All: 20,
			},
		},

		MaxBackups: Parameter[uint64]{
			ParameterCommon: &ParameterCommon{
				ID:                ids.LoggerMaxBackupsID,
				Name:              "Max Archived Logs",
				Description:       "The max number of archived logs to save before deleting old ones.\n\nUse 0 for no limit (preserve all archived logs).",
				AffectsContainers: []ContainerID{ContainerID_Daemon},
			},
			Default: map[Network]uint64{
				Network_All: 3,
			},
		},

		MaxAge: Parameter[uint64]{
			ParameterCommon: &ParameterCommon{
				ID:                ids.LoggerMaxAgeID,
				Name:              "Max Archive Age",
				Description:       "The max number of days an archive log should be preserved for before being deleted.\n\nUse 0 for no limit (preserve all logs regardless of age).",
				AffectsContainers: []ContainerID{ContainerID_Daemon},
			},
			Default: map[Network]uint64{
				Network_All: 90,
			},
		},

		LocalTime: Parameter[bool]{
			ParameterCommon: &ParameterCommon{
				ID:                ids.LoggerLocalTimeID,
				Name:              "Use Local Time",
				Description:       "When a log needs to be archived, by default the system will append the time of archiving to its filename in UTC. Enable this to use your local system's time in the filename instead.",
				AffectsContainers: []ContainerID{ContainerID_Daemon},
			},
			Default: map[Network]bool{
				Network_All: false,
			},
		},

		Compress: Parameter[bool]{
			ParameterCommon: &ParameterCommon{
				ID:                ids.LoggerCompressID,
				Name:              "Compress Archives",
				Description:       "Enable this to compress logs when they get archived to save space.",
				AffectsContainers: []ContainerID{ContainerID_Daemon},
			},
			Default: map[Network]bool{
				Network_All: true,
			},
		},

		EnableHttpTracing: Parameter[bool]{
			ParameterCommon: &ParameterCommon{
				ID:                ids.LoggerEnableHttpTracingID,
				Name:              "Enable HTTP Tracing",
				Description:       "When enabled, each step of every HTTP request for the Execution Client and Beacon Node will be logged. This results in very verbose log files, so only enable this when you need to debug HTTP requests to your clients specifically.",
				AffectsContainers: []ContainerID{ContainerID_Daemon},
			},
			Default: map[Network]bool{
				Network_All: false,
			},
		},
	}
}

// Get the title for the config
func (cfg *LoggerConfig) GetTitle() string {
	return "Logger"
}

// Get the parameters for this config
func (cfg *LoggerConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.Level,
		&cfg.Format,
		&cfg.AddSource,
		&cfg.MaxSize,
		&cfg.MaxBackups,
		&cfg.MaxAge,
		&cfg.LocalTime,
		&cfg.Compress,
		&cfg.EnableHttpTracing,
	}
}

// Get the sections underneath this one
func (cfg *LoggerConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

// Calculate the default number of Geth peers
func (cfg *LoggerConfig) GetOptions() log.LoggerOptions {
	return log.LoggerOptions{
		MaxSize:           int(cfg.MaxSize.Value),
		MaxBackups:        int(cfg.MaxBackups.Value),
		MaxAge:            int(cfg.MaxAge.Value),
		LocalTime:         cfg.LocalTime.Value,
		Compress:          cfg.Compress.Value,
		Format:            cfg.Format.Value,
		Level:             cfg.Level.Value,
		AddSource:         cfg.AddSource.Value,
		EnableHttpTracing: cfg.EnableHttpTracing.Value,
	}
}
