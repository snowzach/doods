package conf

import (
	"net/http"
	"strings"

	config "github.com/spf13/viper"

	"github.com/snowzach/doods/detector/dconfig"
)

func init() {
	// Sets up the config file, environment etc
	config.SetTypeByDefaultValue(true)                      // If a default value is []string{"a"} an environment variable of "a b" will end up []string{"a","b"}
	config.AutomaticEnv()                                   // Automatically use environment variables where available
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Environement variables use underscores instead of periods

	// Logger Defaults
	config.SetDefault("logger.level", "info")
	config.SetDefault("logger.encoding", "console")
	config.SetDefault("logger.color", true)
	config.SetDefault("logger.dev_mode", true)
	config.SetDefault("logger.disable_caller", false)
	config.SetDefault("logger.disable_stacktrace", true)

	// Pidfile
	config.SetDefault("pidfile", "")

	// Profiler config
	config.SetDefault("profiler.enabled", false)
	config.SetDefault("profiler.host", "")
	config.SetDefault("profiler.port", "6060")

	// Server Configuration
	config.SetDefault("server.host", "")
	config.SetDefault("server.port", "8080")
	config.SetDefault("server.tls", false)
	config.SetDefault("server.devcert", false)
	config.SetDefault("server.certfile", "server.crt")
	config.SetDefault("server.keyfile", "server.key")
	config.SetDefault("server.max_msg_size", 64000000)
	config.SetDefault("server.log_requests", true)
	config.SetDefault("server.profiler_enabled", false)
	config.SetDefault("server.profiler_path", "/debug")
	config.SetDefault("server.allowed_origins", []string{"*"})
	config.SetDefault("server.allowed_methods", []string{http.MethodHead, http.MethodOptions, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch})
	config.SetDefault("server.allowed_headers", []string{"*"})
	config.SetDefault("server.allowed_credentials", false)
	config.SetDefault("server.max_age", 300)

	// Main settings
	config.SetDefault("doods.auth_key", "")
	config.SetDefault("doods.detectors", []*dconfig.DetectorConfig{})

}
