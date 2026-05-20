package driver

import (
	"context"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/vpomo/industrial-mcp/internal/domain/entity"
)

type OPCUADriver struct {
	client   *opcua.Client
	endpoint string
	config   map[string]string
}

func NewOPCUADriver() *OPCUADriver {
	return &OPCUADriver{}
}

func (d *OPCUADriver) Type() entity.DataSourceType {
	return entity.DataSourceTypeOPCUA
}

func (d *OPCUADriver) Connect(ctx context.Context, config map[string]string) error {
	endpoint := config["endpoint"]
	if endpoint == "" {
		endpoint = "opc.tcp://localhost:4840"
	}

	client, err := opcua.NewClient(endpoint)
	if err != nil {
		return err
	}

	if err := client.Connect(ctx); err != nil {
		return err
	}

	d.client = client
	d.endpoint = endpoint
	d.config = config
	return nil
}

func (d *OPCUADriver) Disconnect() {
	if d.client != nil {
		d.client.Close(context.Background())
		d.client = nil
	}
}

func (d *OPCUADriver) ReadTag(ctx context.Context, nodeID string) (*entity.Tag, error) {
	if d.client == nil {
		return nil, ErrNotConnected
	}

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		return nil, err
	}

	req := &ua.ReadRequest{
		NodesToRead: []*ua.ReadValueID{{NodeID: id}},
	}

	resp, err := d.client.Read(ctx, req)
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

	return tag, nil
}

func (d *OPCUADriver) WriteTag(ctx context.Context, nodeID string, value interface{}) error {
	if d.client == nil {
		return ErrNotConnected
	}

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		return err
	}

	v, err := ua.NewVariant(value)
	if err != nil {
		return err
	}

	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{{NodeID: id, Value: &ua.DataValue{Value: v}}},
	}

	_, err = d.client.Write(ctx, req)
	return err
}

func (d *OPCUADriver) Scan(ctx context.Context) ([]ScanResult, error) {
	if d.client == nil {
		return nil, ErrNotConnected
	}

	rootID := ua.MustParseNodeID("i=84")
	browseRequest := &ua.BrowseRequest{
		NodesToBrowse: []*ua.BrowseDescription{
			{
				NodeID:          rootID,
				BrowseDirection: ua.BrowseDirectionForward,
				ResultMask:      0x3F,
			},
		},
	}

	resp, err := d.client.Browse(ctx, browseRequest)
	if err != nil {
		return nil, err
	}

	var results []ScanResult
	for _, ref := range resp.Results {
		for _, refDesc := range ref.References {
			results = append(results, ScanResult{
				NodeID:   refDesc.NodeID.String(),
				Name:     refDesc.BrowseName.Name,
				DataType: "unknown",
			})
		}
	}

	return results, nil
}

var ErrNotConnected = &DriverError{Message: "driver not connected"}
var ErrNoResults = &DriverError{Message: "no results returned"}

type DriverError struct {
	Message string
}

func (e *DriverError) Error() string {
	return e.Message
}
