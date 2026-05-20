package opcua

import (
	"context"
	"sync"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/vpomo/industrial-mcp/internal/domain/entity"
)

type OPCUAClient struct {
	client   *opcua.Client
	endpoint string
	tagCache map[string]*entity.Tag
	mu       sync.RWMutex
}

func NewOPCUAClient(endpoint string) (*OPCUAClient, error) {
	client, err := opcua.NewClient(endpoint)
	if err != nil {
		return nil, err
	}
	return &OPCUAClient{
		client:   client,
		endpoint: endpoint,
		tagCache: make(map[string]*entity.Tag),
	}, nil
}

func (c *OPCUAClient) ReadTag(ctx context.Context, nodeID string) (*entity.Tag, error) {
	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		return nil, err
	}

	req := &ua.ReadRequest{
		NodesToRead: []*ua.ReadValueID{
			{NodeID: id},
		},
	}

	resp, err := c.client.Read(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, ErrNoResults
	}

	result := resp.Results[0]
	tag, err := entity.NewTag(nodeID, result.Value.Value())
	if err != nil {
		return nil, err
	}

	if result.Status == ua.StatusGood || result.Status == ua.StatusOK {
		tag.SetQuality(entity.QualityGood)
	} else {
		tag.SetQuality(entity.QualityBad)
	}

	c.mu.Lock()
	c.tagCache[nodeID] = tag
	c.mu.Unlock()

	return tag, nil
}

func (c *OPCUAClient) WriteTag(ctx context.Context, nodeID string, value interface{}) error {
	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		return err
	}

	v, err := ua.NewVariant(value)
	if err != nil {
		return err
	}

	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{NodeID: id, Value: &ua.DataValue{Value: v}},
		},
	}

	_, err = c.client.Write(ctx, req)
	return err
}

func (c *OPCUAClient) Disconnect() {
	if c.client != nil {
		c.client.Close(context.Background())
	}
}

func (c *OPCUAClient) GetCachedTag(nodeID string) (*entity.Tag, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	tag, ok := c.tagCache[nodeID]
	return tag, ok
}

var ErrNoResults = &OpcuaError{Message: "no results returned"}

type OpcuaError struct {
	Message string
}

func (e *OpcuaError) Error() string {
	return e.Message
}
