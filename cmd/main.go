package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/peeley/carpal/internal/config"
	"github.com/peeley/carpal/internal/driver"
	"github.com/peeley/carpal/internal/driver/file"
	"github.com/peeley/carpal/internal/driver/ldap"
	"github.com/peeley/carpal/internal/driver/sql"
	"github.com/peeley/carpal/internal/handler"
)

const (
	DEFAULT_CONFIG_FILE_PATH = "/etc/carpal/config.yml"
	DEFAULT_HTTP_PORT        = "8008"
)

func main() {
	configureLogging()

	fileLocation := os.Getenv("CONFIG_FILE")
	if fileLocation == "" {
		slog.Debug(
			fmt.Sprintf(
				"no config file specified, using default config file path %s",
				DEFAULT_CONFIG_FILE_PATH,
			),
		)
		fileLocation = DEFAULT_CONFIG_FILE_PATH
	}

	expandEnvs := os.Getenv("EXPAND_CONFIG_ENV_VARS") != ""

	configWizard := config.NewConfigWizard(fileLocation, expandEnvs)
	config, err := configWizard.GetConfiguration()
	if err != nil {
		slog.Error("could not load configuration", "err", err)
		os.Exit(1)
	}

	var driver driver.Driver
	switch config.Driver {
	case "file":
		driver = file.NewFileDriver(*config)
	case "ldap":
		driver = ldap.NewLDAPDriver(*config)
	case "sql":
		var err error
		driver, err = sql.NewSQLDriver(*config)
		if err != nil {
			slog.Error("failed to initialize SQL driver", "err", err)
			os.Exit(1)
		}
	default:
		slog.Error(fmt.Sprintf("driver `%s` is invalid", config.Driver))
		os.Exit(1)
	}

	handler := handler.NewResourceHandler(driver)
	http.HandleFunc("/", handler.Handle)

	port := os.Getenv("PORT")
	if port == "" {
		slog.Debug(
			fmt.Sprintf(
				"no http port specified, using default port number %s",
				DEFAULT_HTTP_PORT,
			),
		)
		port = DEFAULT_HTTP_PORT
	}

	slog.Info(fmt.Sprintf("launching carpal server on port %v...", port))
	slog.Error(fmt.Sprintf("%v", http.ListenAndServe(":"+port, nil)))
}

func configureLogging() {
	logLevels := map[string]slog.Level{
		"DEBUG":   slog.LevelDebug,
		"INFO":    slog.LevelInfo,
		"WARNING": slog.LevelWarn,
		"ERROR":   slog.LevelError,
	}

	envLogLevel := os.Getenv("LOG_LEVEL")
	logLevel, ok := logLevels[envLogLevel]
	if !ok {
		logLevel = slog.LevelInfo
	}

	slog.SetLogLoggerLevel(logLevel)
}
