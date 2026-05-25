package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/vpomo/industrial-mcp/internal/application/command"
	"github.com/vpomo/industrial-mcp/internal/application/query"
	"github.com/vpomo/industrial-mcp/internal/domain/service"
	"github.com/vpomo/industrial-mcp/internal/infrastructure/mqtt"
	infrarepo "github.com/vpomo/industrial-mcp/internal/infrastructure/repository"
	"github.com/vpomo/industrial-mcp/internal/interfaces/mcp"
	"github.com/vpomo/industrial-mcp/pkg/license"
	"github.com/vpomo/industrial-mcp/pkg/logger"
	"github.com/vpomo/industrial-mcp/pkg/x402"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	MQTT    MQTTConfig    `yaml:"mqtt"`
	OPCUA   OPCUAConfig   `yaml:"opcua"`
	License LicenseConfig `yaml:"license"`
	X402    X402Config    `yaml:"x402"`
	Metrics MetricsConfig `yaml:"metrics"`
	Logging LoggingConfig `yaml:"logging"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type MQTTConfig struct {
	BrokerURL   string `yaml:"broker_url"`
	ClientID    string `yaml:"client_id"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	TopicPrefix string `yaml:"topic_prefix"`
	QoS         int    `yaml:"qos"`
	Retain      bool   `yaml:"retain"`
}

type OPCUAConfig struct {
	Endpoint     string `yaml:"endpoint"`
	SecurityMode string `yaml:"security_mode"`
	CertFile     string `yaml:"cert_file"`
	KeyFile      string `yaml:"key_file"`
}

type LicenseConfig struct {
	Enabled       bool   `yaml:"enabled"`
	PublicKeyPath string `yaml:"public_key_path"`
}

type X402Config struct {
	Enabled        bool   `yaml:"enabled"`
	PaymentAddress string `yaml:"payment_address"`
}

type MetricsConfig struct {
	Enabled    bool   `yaml:"enabled"`
	File       string `yaml:"file"`
	BufferSize int    `yaml:"buffer_size"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

const licenseCheckInterval = 20 * time.Minute

func main() {
	configPath := flag.String("config", "cmd/server/config.yaml", "path to config file")
	flag.Parse()

	absConfigPath, err := filepath.Abs(*configPath)
	if err != nil {
		log.Fatalf("failed to resolve config path: %v", err)
	}

	cfg, err := loadConfig(absConfigPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger := logger.New(cfg.Logging.Level)
	appLogger.Info("starting MCP server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tagRepo := infrarepo.NewMemoryTagRepository()

	metricsRepo, err := infrarepo.NewMemoryMetricsRepository(cfg.Metrics.File)
	if err != nil {
		appLogger.Warn("metrics disabled", "error", err.Error())
		metricsRepo = nil
	}

	tagService := service.NewTagService(tagRepo)

	mqttCfg := mqttConfigFromApp(cfg)
	applyMQTTEnv(&mqttCfg)

	var mqttPublisher mqtt.PublishPublisher
	var mqttSubscriber mqtt.TagSubscriber
	var mqttLazy *mqtt.LazyClient
	if mqttCfg.Enabled() {
		mqttLazy = mqtt.NewLazyClient(mqttCfg)
		mqttPublisher = mqttLazy
		mqttSubscriber = mqttLazy
		appLogger.Info("MQTT enabled", "broker", mqttCfg.BrokerURL, "prefix", mqttCfg.TopicPrefix)
	} else {
		appLogger.Warn("MQTT disabled", "reason", "empty broker_url")
	}

	readTagH := query.NewReadTagHandler(tagService)

	writeTagH := command.NewWriteTagHandler(tagRepo, mqttPublisher)
	subTagH := command.NewSubscribeTagHandler(mqttSubscriber)

	mcpServerCfg := &mcp.Config{
		ListenAddr:           cfg.Server.Host + ":" + itoa(cfg.Server.Port),
		LogLevel:             cfg.Logging.Level,
		MetricsFile:          cfg.Metrics.File,
		X402Enabled:          cfg.X402.Enabled,
		X402PaymentAddress:   cfg.X402.PaymentAddress,
		LicensePublicKeyPath: cfg.License.PublicKeyPath,
	}

	server := mcp.NewMCPServer(
		mcpServerCfg,
		nil,
		readTagH,
		writeTagH,
		subTagH,
		metricsRepo,
	)

	if cfg.License.Enabled {
		licenseFile := licensePathBesideConfig(absConfigPath)

		var publicKeyPEM []byte
		if cfg.License.PublicKeyPath != "" {
			publicKeyPath := resolvePathBesideConfig(absConfigPath, cfg.License.PublicKeyPath)
			publicKeyPEM, err = os.ReadFile(publicKeyPath)
			if err != nil {
				log.Fatalf("failed to read license public key %s: %v", publicKeyPath, err)
			}
		}

		lv, err := license.New(
			publicKeyPEM,
			licenseFile,
			license.WithCheckInterval(licenseCheckInterval),
		)
		if err != nil {
			log.Fatalf("license system error: %v", err)
		}

		server.SetLicenseValidator(lv)

		if err := lv.Validate(); err != nil {
			appLogger.Warn(
				"license validation failed at startup; MCP requests will be rejected until license is fixed",
				"file", licenseFile,
				"error", err.Error(),
			)
		} else {
			appLogger.Info("license validated", "file", licenseFile)
			go runPeriodicLicenseCheck(ctx, cancel, appLogger, lv, licenseFile, licenseCheckInterval)
		}
	}

	if cfg.X402.Enabled {
		x402Handler := x402.NewHandler(cfg.X402.Enabled, cfg.X402.PaymentAddress)
		server.SetX402Handler(x402Handler)
	}

	ready := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start(ctx, ready)
	}()

	select {
	case err := <-errCh:
		log.Fatalf(
			"failed to start server on %s: %v\nstop the process using this port (GoLand: red Stop button, or: fuser -k %d/tcp)",
			mcpServerCfg.ListenAddr,
			err,
			cfg.Server.Port,
		)
	case <-ready:
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
	case err := <-errCh:
		if err != nil {
			appLogger.Error("server error", "error", err.Error())
		}
	}

	appLogger.Info("shutting down server")
	if mqttLazy != nil {
		mqttLazy.Disconnect()
	}
	cancel()

	time.Sleep(time.Second)
}

func runPeriodicLicenseCheck(
	ctx context.Context,
	cancel context.CancelFunc,
	appLogger *logger.Logger,
	lv *license.Validator,
	licenseFile string,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := lv.Validate(); err != nil {
				appLogger.Error(
					"periodic license validation failed",
					"file", licenseFile,
					"error", err.Error(),
				)
				cancel()
				return
			}
			appLogger.Info("periodic license validation ok", "file", licenseFile)
		case <-ctx.Done():
			return
		}
	}
}

func licensePathBesideConfig(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "license.dat")
}

func resolvePathBesideConfig(configPath, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(filepath.Dir(configPath), p)
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

func mqttConfigFromApp(cfg *Config) mqtt.Config {
	qos := byte(cfg.MQTT.QoS)
	if qos > 2 {
		qos = 0
	}
	return mqtt.Config{
		BrokerURL:   cfg.MQTT.BrokerURL,
		ClientID:    cfg.MQTT.ClientID,
		Username:    cfg.MQTT.Username,
		Password:    cfg.MQTT.Password,
		TopicPrefix: cfg.MQTT.TopicPrefix,
		QoS:         qos,
		Retain:      cfg.MQTT.Retain,
	}
}

func applyMQTTEnv(cfg *mqtt.Config) {
	if v := os.Getenv("MQTT_BROKER_URL"); v != "" {
		cfg.BrokerURL = v
	}
	if v := os.Getenv("MQTT_CLIENT_ID"); v != "" {
		cfg.ClientID = v
	}
	if v := os.Getenv("MQTT_USERNAME"); v != "" {
		cfg.Username = v
	}
	if v := os.Getenv("MQTT_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("MQTT_TOPIC_PREFIX"); v != "" {
		cfg.TopicPrefix = v
	}
	if v := os.Getenv("MQTT_RETAIN"); v != "" {
		retain, err := strconv.ParseBool(v)
		if err == nil {
			cfg.Retain = retain
		}
	}
}
