package mqtt

import "testing"

func TestConfigRetainPassedToClient(t *testing.T) {
	client := &MQTTClient{retain: true}
	if !client.retain {
		t.Fatal("expected retain=true on client")
	}

	cfg := Config{Retain: true}
	if !cfg.Retain {
		t.Fatal("expected retain=true in config")
	}
}
