// xk6 build --with github.com/grafana/xk6-ethereum=.
package ethereum

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/grafana/sobek"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

const (
	privateKey = "42b6e34dc21598a807dc19d7784c71b2a7a01f6480dc6f58258f78e539f1a1fa"
)

type ethMetrics struct {
	RequestDuration *metrics.Metric
	TimeToMine      *metrics.Metric
	Block           *metrics.Metric
	GasUsed         *metrics.Metric
	TPS             *metrics.Metric
	BlockTime       *metrics.Metric
}

func init() {
	modules.Register("k6/x/ethereum", &EthRoot{})
}

// EthRoot is the root module
type EthRoot struct{}

// NewModuleInstance implements the modules.Module interface returning a new instance for each VU.
func (*EthRoot) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ModuleInstance{
		vu: vu,
		m:  registerMetrics(vu),
	}
}

type ModuleInstance struct {
	vu modules.VU
	m  ethMetrics
}

// Exports implements the modules.Instance interface and returns the exported types for the JS module.
func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{Named: map[string]interface{}{
		"Client": mi.NewClient,
	}}
}

func (mi *ModuleInstance) NewClient(call sobek.ConstructorCall) *sobek.Object {
	rt := mi.vu.Runtime()

	var optionsArg map[string]interface{}
	err := rt.ExportTo(call.Arguments[0], &optionsArg)
	if err != nil {
		common.Throw(rt, errors.New("unable to parse options object"))
	}

	opts, err := newOptionsFrom(optionsArg)
	if err != nil {
		common.Throw(rt, fmt.Errorf("invalid options; reason: %w", err))
	}

	if opts.URL == "" {
		opts.URL = "http://localhost:8545"
	}

	if opts.PrivateKey == "" {
		opts.PrivateKey = privateKey
	}

	var wa *wallet.Key
	if opts.Mnemonic != "" {
		w, err := wallet.NewWalletFromMnemonic(opts.Mnemonic)
		if err != nil {
			common.Throw(rt, fmt.Errorf("invalid options; reason: %w", err))
		}
		wa = w
	} else if opts.PrivateKey != "" {
		pk, err := hex.DecodeString(opts.PrivateKey)
		if err != nil {
			common.Throw(rt, fmt.Errorf("invalid options; reason: %w", err))
		}
		w, err := wallet.NewWalletFromPrivKey(pk)
		if err != nil {
			common.Throw(rt, fmt.Errorf("invalid options; reason: %w", err))
		}
		wa = w
	}

	c, err := jsonrpc.NewClient(opts.URL)
	if err != nil {
		common.Throw(rt, fmt.Errorf("invalid options; reason: %w", err))
	}

	cid, err := c.Eth().ChainID()
	if err != nil {
		common.Throw(rt, fmt.Errorf("invalid options; reason: %w", err))
	}

	client := &Client{
		vu:      mi.vu,
		metrics: mi.m,
		client:  c,
		w:       wa,
		chainID: cid,
		opts:    opts,
	}

	go client.pollForBlocks()

	return rt.ToValue(client).ToObject(rt)
}

func registerMetrics(vu modules.VU) ethMetrics {
	registry := vu.InitEnv().Registry
	m := ethMetrics{
		RequestDuration: registry.MustNewMetric("ethereum_req_duration", metrics.Trend, metrics.Time),
		TimeToMine:      registry.MustNewMetric("ethereum_time_to_mine", metrics.Trend, metrics.Time),
		Block:           registry.MustNewMetric("ethereum_block", metrics.Counter, metrics.Default),
		GasUsed:         registry.MustNewMetric("ethereum_gas_used", metrics.Trend, metrics.Default),
		TPS:             registry.MustNewMetric("ethereum_tps", metrics.Trend, metrics.Default),
		BlockTime:       registry.MustNewMetric("ethereum_block_time", metrics.Trend, metrics.Time),
	}

	return m
}

func (c *Client) reportMetricsFromStats(call string, t time.Duration) {
	registry := metrics.NewRegistry()
	metrics.PushIfNotDone(c.vu.Context(), c.vu.State().Samples, metrics.Sample{
		TimeSeries: metrics.TimeSeries{
			Metric: c.metrics.RequestDuration,
			Tags:   registry.RootTagSet().With("call", call),
		},
		Value: float64(t / time.Millisecond),
		Time:  time.Now(),
	})
}

// options defines configuration options for the client.
type options struct {
	URL        string `json:"url,omitempty"`
	Mnemonic   string `json:"mnemonic,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
}

// newOptionsFrom validates and instantiates an options struct from its map representation
// as obtained by calling a Goja's Runtime.ExportTo.
func newOptionsFrom(argument map[string]interface{}) (*options, error) {
	jsonStr, err := json.Marshal(argument)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize options to JSON %w", err)
	}

	// Instantiate a JSON decoder which will error on unknown
	// fields. As a result, if the input map contains an unknown
	// option, this function will produce an error.
	decoder := json.NewDecoder(bytes.NewReader(jsonStr))
	decoder.DisallowUnknownFields()

	var opts options
	err = decoder.Decode(&opts)
	if err != nil {
		return nil, fmt.Errorf("unable to decode options %w", err)
	}

	return &opts, nil
}
