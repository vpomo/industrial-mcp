package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imatic/mcp_mqtt_opcua/internal/application/command"
	"github.com/imatic/mcp_mqtt_opcua/internal/application/query"
	"github.com/imatic/mcp_mqtt_opcua/internal/domain/service"
	"github.com/imatic/mcp_mqtt_opcua/internal/infrastructure/mqtt"
	infrarepo "github.com/imatic/mcp_mqtt_opcua/internal/infrastructure/repository"
	"github.com/imatic/mcp_mqtt_opcua/internal/interfaces/mcp"
	"github.com/imatic/mcp_mqtt_opcua/pkg/license"
	"github.com/imatic/mcp_mqtt_opcua/pkg/x402"
	"github.com/imatic/mcp_mqtt_opcua/pkg/logger"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	MQTT     MQTTConfig    `yaml:"mqtt"`
	OPCUA    OPCUAConfig   `yaml:"opcua"`
	License  LicenseConfig  `yaml:"license"`
	X402     X402Config    `yaml:"x402"`
	Metrics  MetricsConfig `yaml:"metrics"`
	Logging  LoggingConfig `yaml:"logging"`
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
}

type OPCUAConfig struct {
	Endpoint     string `yaml:"endpoint"`
	SecurityMode string `yaml:"security_mode"`
	CertFile     string `yaml:"cert_file"`
	KeyFile      string `yaml:"key_file"`
}

type LicenseConfig struct {
	Enabled        bool   `yaml:"enabled"`
	PublicKeyPath  string `yaml:"public_key_path"`
	CheckInterval  int    `yaml:"check_interval"`
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

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger := logger.New(cfg.Logging.Level)
	appLogger.Info("starting MCP server")

	tagRepo := infrarepo.NewMemoryTagRepository()

	metricsRepo, err := infrarepo.NewMemoryMetricsRepository(cfg.Metrics.File)
	if err != nil {
		appLogger.Warn("metrics disabled", "error", err.Error())
		metricsRepo = nil
	}

	tagService := service.NewTagService(tagRepo)

	var mqttClient *mqtt.MQTTClient
	if cfg.MQTT.BrokerURL != "" {
		mqttClient, err = mqtt.NewMQTTClient(
			cfg.MQTT.BrokerURL,
			cfg.MQTT.ClientID,
			cfg.MQTT.TopicPrefix,
		)
		if err != nil {
			appLogger.Warn("MQTT disabled", "error", err.Error())
		}
	}

	readTagH := query.NewReadTagHandler(tagService)
	writeTagH := command.NewWriteTagHandler(tagRepo, mqttClient)
	subTagH := command.NewSubscribeTagHandler(mqttClient)

	mcpServerCfg := &mcp.Config{
		ListenAddr:            cfg.Server.Host + ":" + itoa(cfg.Server.Port),
		LogLevel:              cfg.Logging.Level,
		MetricsFile:           cfg.Metrics.File,
		X402Enabled:           cfg.X402.Enabled,
		X402PaymentAddress:    cfg.X402.PaymentAddress,
		LicensePublicKeyPath:  cfg.License.PublicKeyPath,
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
		lv, err := license.New(cfg.License.PublicKeyPath)
		if err != nil {
			appLogger.Warn("license system error", "error", err.Error())
		} else {
			server.SetLicenseValidator(lv)
		}
	}

	if cfg.X402.Enabled {
		x402Handler := x402.NewHandler(cfg.X402.Enabled, cfg.X402.PaymentAddress)
		server.SetX402Handler(x402Handler)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil {
			appLogger.Error("server error", "error", err.Error())
			cancel()
		}
	}()

	appLogger.Info("server started", "addr", mcpServerCfg.ListenAddr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down server")
	cancel()

	time.Sleep(time.Second)
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