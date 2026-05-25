package mqtt

type Config struct {
	BrokerURL   string
	ClientID    string
	Username    string
	Password    string
	TopicPrefix string
	QoS         byte
	Retain      bool
}

func (c Config) Enabled() bool {
	return c.BrokerURL != ""
}
