package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/frangi01/bbtelgo/internal/logx"
	"github.com/joho/godotenv"
)

type Mode string

const (
	ModeWebhook Mode = "webhook"
	ModePolling Mode = "polling"
)

type LogLevel string

const (
	LogLevelDebug 	LogLevel = "debug"
	LogLevelInfo 	LogLevel = "info"
	LogLevelError	LogLevel = "error"
)

type MongoCfg struct {
	URI           string
	DB            string
	MinPoolSize   uint64
	MaxPoolSize   uint64
	AppName       string
	ConnectTimeout time.Duration
	CmdTimeout     time.Duration
}


type Config struct {
	LogLevel					LogLevel
	LogFile						bool
	LogFilePath					string
	Mode						Mode
	Token						string
	ResetWebHook 				bool
	Timeout						int
	TransportMaxIdleConns		int
	TransportIdleConnTimeout	int
	WebHookSecret				string
	WebHookPort 				string
	WebHookPublicUrl 			string
	WebHookTLSKeyFile 			string
	WebHookTLSCertFile 			string
	MongoCfg					MongoCfg
}

func Load(logger *logx.Logger) (Config, error) {
	_ = godotenv.Load()

	mode := Mode(os.Getenv("APP_MODE"))
	logLevel := LogLevel(os.Getenv("APP_LOG_LEVEL"))
	
	strTimeout := os.Getenv("APP_HTTPCLIENT_TIMEOUT")
	timeout, err := strconv.Atoi(strTimeout)
	if err != nil {
		logger.Errorf("env APP_HTTPCLIENT_TIMEOUT")
	}

	strTransportMaxIdleConns := os.Getenv("APP_HTTPCLIENT_TRANSPORT_MAXIDLECONNS")
	transportMaxIdleConns, err := strconv.Atoi(strTransportMaxIdleConns)
	if err != nil {
		logger.Errorf("env APP_HTTPCLIENT_TRANSPORT_MAXIDLECONNS")
	}
	
	strTransportIdleConnTimeout := os.Getenv("APP_HTTPCLIENT_TRANSPORT_IDLECONNTIMEOUT")
	transportIdleConnTimeout, err := strconv.Atoi(strTransportIdleConnTimeout)
	if err != nil {
		logger.Errorf("env APP_HTTPCLIENT_TRANSPORT_IDLECONNTIMEOUT")
	}

	strMongoMinPoolSize := os.Getenv("MONGO_MIN_POOL_SIZE")
	mongoMinPoolSize, err := strconv.ParseUint(strMongoMinPoolSize, 10, 32)
	if err != nil {
		logger.Errorf("env MONGO_MIN_POOL_SIZE")
	}

	strMongoMaxPoolSize := os.Getenv("MONGO_MAX_POOL_SIZE")
	mongoMaxPoolSize, err := strconv.ParseUint(strMongoMaxPoolSize, 10, 32)
	if err != nil {
		logger.Errorf("env MONGO_MAX_POOL_SIZE")
	}

	strMongoConnectTimeout := os.Getenv("MONGO_CONNECTION_TIMEOUT")
	mongoConnectTimeout, err := strconv.Atoi(strMongoConnectTimeout)
	if err != nil {
		logger.Errorf("env MONGO_CONNECTION_TIMEOUT")
	}

	strMongoCmdTimeout := os.Getenv("MONGO_CMD_TIMEOUT")
	mongoCmdTimeout, err := strconv.Atoi(strMongoCmdTimeout)
	if err != nil {
		logger.Errorf("env MONGO_CMD_TIMEOUT")
	}

	cfg := Config{
		LogLevel: 					logLevel,
		LogFile:					os.Getenv("APP_LOG_FILE") == "true",
		LogFilePath:				os.Getenv("APP_LOG_FILE_PATH"),
		Mode: 						mode,
		Token: 						os.Getenv("APP_TELEGRAM_TOKEN"),
		ResetWebHook:				os.Getenv("APP_RESET_WEBHOOK") == "true",
		Timeout:  					timeout,
		TransportMaxIdleConns: 		transportMaxIdleConns,
		TransportIdleConnTimeout: 	transportIdleConnTimeout,
		WebHookSecret: os.Getenv("APP_WEBHOOK_SECRET"),
		WebHookPublicUrl: os.Getenv("APP_WEBHOOK_PUBLIC_URL"),
		WebHookPort: os.Getenv("APP_WEBHOOK_PORT"),
		WebHookTLSCertFile: os.Getenv("APP_WEBHOOK_TLS_CERT_FILE"),
		WebHookTLSKeyFile: os.Getenv("APP_WEBHOOK_TLS_KEY_FILE"),
		MongoCfg: MongoCfg{
			URI: os.Getenv("MONGO_URI"),
			DB: os.Getenv("MONGO_DB"),
			MinPoolSize: mongoMinPoolSize,
			MaxPoolSize: mongoMaxPoolSize,
			AppName: os.Getenv("MONGO_APPNAME"),
			ConnectTimeout: time.Duration(mongoConnectTimeout),
			CmdTimeout: time.Duration(mongoCmdTimeout),
		},
	}

	if cfg.Token == "" {
		err := fmt.Errorf("APP_TELEGRAM_TOKEN missing")
		logger.Errorf("%v", err)
		return Config{}, err
	}

	if cfg.Mode != ModeWebhook && cfg.Mode != ModePolling {
		err := fmt.Errorf("APP_MODE should be 'webhook' or 'polling'")
		logger.Errorf("%v", err)
		return Config{}, err
	}

	if cfg.Mode == ModeWebhook {
		var missing []string
		if cfg.WebHookPort == "" {
			missing = append(missing, "APP_WEBHOOK_PORT")
		}
		if cfg.WebHookPublicUrl == "" {
			missing = append(missing, "APP_WEBHOOK_PUBLIC_URL")
		}
		if cfg.WebHookTLSCertFile == "" {
			missing = append(missing, "APP_WEBHOOK_TLS_CERT_FILE")
		}
		if cfg.WebHookTLSKeyFile == "" {
			missing = append(missing, "APP_WEBHOOK_TLS_KEY_FILE")
		}

		if len(missing) > 0 {
			err := fmt.Errorf("in webhook mode you have to set: %v", missing)
			logger.Errorf("%v", err)
			return Config{}, err
		}

	}

	return cfg, nil
}