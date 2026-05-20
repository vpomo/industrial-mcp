package opcua

import (
	"testing"

	"github.com/vpomo/mcp_mqtt_opcua/internal/domain/entity"
)

func TestOPCUAClientTagCache(t *testing.T) {
	client := &OPCUAClient{
		endpoint: "opc.tcp://localhost:4840",
		tagCache: make(map[string]*entity.Tag),
	}

	tag, _ := entity.NewTag("ns=2;i=1", 42.0)
	client.tagCache["ns=2;i=1"] = tag

	cached, ok := client.GetCachedTag("ns=2;i=1")
	if !ok {
		t.Fatal("expected tag to be cached")
	}
	if cached.Value() != 42.0 {
		t.Errorf("expected value 42.0, got %v", cached.Value())
	}

	_, ok = client.GetCachedTag("nonexistent")
	if ok {
		t.Error("expected not found for nonexistent node")
	}
}

func TestOpcuaError(t *testing.T) {
	err := &OpcuaError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got %s", err.Error())
	}
}

func TestErrNoResults(t *testing.T) {
	if ErrNoResults.Error() != "no results returned" {
		t.Errorf("unexpected error message: %s", ErrNoResults.Error())
	}
}
