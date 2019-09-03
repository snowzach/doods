package conf

import (
	"github.com/blendle/zapdriver"
	config "github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() {

	logConfig := zap.NewProductionConfig()

	// Log Level
	var logLevel zapcore.Level
	if err := logLevel.Set(config.GetString("logger.level")); err != nil {
		zap.S().Fatalw("Could not determine logger.level", "error", err)
	}
	logConfig.Level.SetLevel(logLevel)

	// Handle different logger encodings
	loggerEncoding := config.GetString("logger.encoding")
	switch loggerEncoding {
	case "stackdriver":
		logConfig.Encoding = "json"
		logConfig.EncoderConfig = zapdriver.NewDevelopmentEncoderConfig()
	default:
		logConfig.Encoding = loggerEncoding
		// Enable Color
		if config.GetBool("logger.color") {
			logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		logConfig.DisableStacktrace = config.GetBool("logger.disable_stacktrace")
		// Use sane timestamp when logging to console
		if logConfig.Encoding == "console" {
			logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		}

		// JSON Fields
		logConfig.EncoderConfig.MessageKey = "msg"
		logConfig.EncoderConfig.LevelKey = "level"
		logConfig.EncoderConfig.CallerKey = "caller"
	}

	// Settings
	logConfig.Development = config.GetBool("logger.dev_mode")
	logConfig.DisableCaller = config.GetBool("logger.disable_caller")

	// Build the logger
	globalLogger, _ := logConfig.Build()
	zap.ReplaceGlobals(globalLogger)

}
