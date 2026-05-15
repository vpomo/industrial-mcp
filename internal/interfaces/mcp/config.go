package mcp

type Config struct {
	ListenAddr            string
	LogLevel              string
	MetricsFile           string
	X402Enabled           bool
	X402PaymentAddress    string
	LicensePublicKeyPath  string
}

func NewConfig() *Config {
	return &Config{
		ListenAddr:  "0.0.0.0:8080",
		LogLevel:    "info",
		MetricsFile: "/app/logs/requests.jsonl",
		X402Enabled: false,
	}
}