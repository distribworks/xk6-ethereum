// xk6 build --with github.com/distribworks/xk6-ethereum=.
package ethereum

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/grafana/sobek"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
	"github.com/umbracle/ethgo/contract"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

type Transaction struct {
	From      string
	To        string
	Input     []byte
	GasPrice  uint64
	GasFeeCap uint64
	GasTipCap uint64
	Gas       uint64
	Value     int64
	Nonce     uint64
	// eip-2930 values
	ChainId int64
}

type Client struct {
	w       *wallet.Key
	client  *jsonrpc.Client
	chainID *big.Int
	vu      modules.VU
	metrics ethMetrics
	opts    *options
}

func (c *Client) Exports() modules.Exports {
	return modules.Exports{}
}

func (c *Client) Call(method string, params ...interface{}) (interface{}, error) {
	t := time.Now()
	var out interface{}
	err := c.client.Call(method, &out, params...)
	c.reportMetricsFromStats(method, time.Since(t))
	return out, err
}

func (c *Client) GasPrice() (uint64, error) {
	t := time.Now()
	g, err := c.client.Eth().GasPrice()
	c.reportMetricsFromStats("gas_price", time.Since(t))
	return g, err
}

func (c *Client) GetBalance(address string, blockNumber ethgo.BlockNumber) (uint64, error) {
	b, err := c.client.Eth().GetBalance(ethgo.HexToAddress(address), blockNumber)
	return b.Uint64(), err
}

// BlockNumber returns the current block number.
func (c *Client) BlockNumber() (uint64, error) {
	return c.client.Eth().BlockNumber()
}

// GetBlockByNumber returns the block with the given block number.
func (c *Client) GetBlockByNumber(number ethgo.BlockNumber, full bool) (*ethgo.Block, error) {
	return c.client.Eth().GetBlockByNumber(number, full)
}

// GetNonce returns the nonce for the given address.
func (c *Client) GetNonce(address string) (uint64, error) {
	return c.client.Eth().GetNonce(ethgo.HexToAddress(address), ethgo.Pending)
}

// EstimateGas returns the estimated gas for the given transaction.
func (c *Client) EstimateGas(tx Transaction) (uint64, error) {
	to := ethgo.HexToAddress(tx.To)

	msg := &ethgo.CallMsg{
		From:     c.w.Address(),
		To:       &to,
		Value:    big.NewInt(tx.Value),
		Data:     tx.Input,
		GasPrice: tx.GasPrice,
	}

	gas, err := c.client.Eth().EstimateGas(msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %e", err)
	}

	return gas, nil
}

// SendTransaction sends a transaction to the network.
func (c *Client) SendTransaction(tx Transaction) (string, error) {
	to := ethgo.HexToAddress(tx.To)

	if tx.Gas == 0 {
		tx.Gas = 21000
	}

	if tx.GasPrice == 0 && tx.GasFeeCap == 0 && tx.GasTipCap == 0 {
		tx.GasPrice = 5242880
	}

	t := &ethgo.Transaction{
		Type:     ethgo.TransactionLegacy,
		From:     ethgo.HexToAddress(tx.From),
		To:       &to,
		Value:    big.NewInt(tx.Value),
		Gas:      tx.Gas,
		GasPrice: tx.GasPrice,
	}

	if tx.GasFeeCap > 0 || tx.GasTipCap > 0 {
		t.Type = ethgo.TransactionDynamicFee
		t.GasPrice = 0
		t.MaxFeePerGas = big.NewInt(0).SetUint64(tx.GasFeeCap)
		t.MaxPriorityFeePerGas = big.NewInt(0).SetUint64(tx.GasTipCap)
	}

	h, err := c.client.Eth().SendTransaction(t)
	return h.String(), err
}

// SendRawTransaction signs and sends transaction to the network.
func (c *Client) SendRawTransaction(tx Transaction) (string, error) {
	to := ethgo.HexToAddress(tx.To)

	gas, err := c.EstimateGas(tx)
	if err != nil {
		return "", err
	}

	t := &ethgo.Transaction{
		Type:     ethgo.TransactionLegacy,
		From:     c.w.Address(),
		To:       &to,
		Value:    big.NewInt(tx.Value),
		Gas:      gas,
		GasPrice: tx.GasPrice,
		Nonce:    tx.Nonce,
		Input:    tx.Input,
		ChainID:  c.chainID,
	}

	if tx.GasFeeCap > 0 || tx.GasTipCap > 0 {
		t.Type = ethgo.TransactionDynamicFee
		t.GasPrice = 0
		t.MaxFeePerGas = big.NewInt(0).SetUint64(tx.GasFeeCap)
		t.MaxPriorityFeePerGas = big.NewInt(0).SetUint64(tx.GasTipCap)
	}

	s := wallet.NewEIP155Signer(t.ChainID.Uint64())
	st, err := s.SignTx(t, c.w)
	if err != nil {
		return "", err
	}

	trlp, err := st.MarshalRLPTo(nil)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tx: %e", err)
	}

	h, err := c.client.Eth().SendRawTransaction(trlp)
	return h.String(), err
}

// GetTransactionReceipt returns the transaction receipt for the given transaction hash.
func (c *Client) GetTransactionReceipt(hash string) (*ethgo.Receipt, error) {
	r, err := c.client.Eth().GetTransactionReceipt(ethgo.HexToHash(hash))
	if err != nil {
		return nil, err
	}

	if r != nil {
		return r, nil
	}

	return nil, fmt.Errorf("not found")
}

// WaitForTransactionReceipt waits for the transaction receipt for the given transaction hash.
func (c *Client) WaitForTransactionReceipt(hash string) *sobek.Promise {
	promise, resolve, reject := c.makeHandledPromise()
	now := time.Now()

	go func() {
		for {
			receipt, err := c.GetTransactionReceipt(hash)
			if err != nil {
				if err.Error() != "not found" {
					reject(err)
					return
				}
			}
			if receipt != nil {
				// If we are testing vu is nil
				if c.vu != nil {
					// Report metrics
					metrics.PushIfNotDone(c.vu.Context(), c.vu.State().Samples, metrics.Sample{
						TimeSeries: metrics.TimeSeries{
							Metric: c.metrics.TimeToMine,
							Tags:   metrics.NewRegistry().RootTagSet(),
						},
						Value: float64(time.Since(now) / time.Millisecond),
						Time:  time.Now(),
					})
				}
				resolve(receipt)
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return promise
}

// Accounts returns a list of addresses owned by client. This endpoint is not enabled in infrastructure providers.
func (c *Client) Accounts() ([]string, error) {
	accounts, err := c.client.Eth().Accounts()
	if err != nil {
		return nil, err
	}

	addresses := make([]string, len(accounts))
	for i, a := range accounts {
		addresses[i] = a.String()
	}

	return addresses, nil
}

// NewContract creates a new contract instance with the given ABI.
func (c *Client) NewContract(address string, abistr string) (*Contract, error) {
	contractABI, err := abi.NewABI(abistr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse abi: %w", err)
	}

	opts := []contract.ContractOption{
		contract.WithJsonRPC(c.client.Eth()),
		contract.WithSender(c.w),
	}

	contract := contract.NewContract(ethgo.HexToAddress(address), contractABI, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}

	return &Contract{
		Contract: contract,
		client:   c,
	}, nil
}

// DeployContract deploys a contract to the blockchain.
func (c *Client) DeployContract(abistr string, bytecode string, args ...interface{}) (*ethgo.Receipt, error) {
	// Parse ABI
	contractABI, err := abi.NewABI(abistr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse abi: %w", err)
	}

	// Parse bytecode
	contractBytecode, err := hex.DecodeString(bytecode)
	if err != nil {
		return nil, fmt.Errorf("failed to decode bytecode: %w", err)
	}

	opts := []contract.ContractOption{
		contract.WithJsonRPC(c.client.Eth()),
		contract.WithSender(c.w),
	}

	// Deploy contract
	txn, err := contract.DeployContract(contractABI, contractBytecode, args, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy contract: %w", err)
	}
	txn.WithOpts(&contract.TxnOpts{
		GasLimit: 1500000,
	})

	err = txn.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to deploy contract: %w", err)
	}

	receipt, err := txn.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed waiting to deploy contract: %w", err)
	}

	return receipt, nil
}

// makeHandledPromise will create a promise and return its resolve and reject methods,
// wrapped in such a way that it will block the eventloop from exiting before they are
// called even if the promise isn't resolved by the time the current script ends executing.
func (c *Client) makeHandledPromise() (*sobek.Promise, func(interface{}), func(interface{})) {
	runtime := c.vu.Runtime()
	callback := c.vu.RegisterCallback()
	p, resolve, reject := runtime.NewPromise()

	return p, func(i interface{}) {
			// more stuff
			callback(func() error {
				resolve(i)
				return nil
			})
		}, func(i interface{}) {
			// more stuff
			callback(func() error {
				reject(i)
				return nil
			})
		}
}

var blocks sync.Map

// PollBlocks polls for new blocks and emits a "block" metric.
func (c *Client) pollForBlocks() {
	var lastBlockNumber uint64
	var prevBlock *ethgo.Block

	now := time.Now()

	for range time.Tick(500 * time.Millisecond) {
		blockNumber, err := c.BlockNumber()
		if err != nil {
			panic(err)
		}

		if blockNumber > lastBlockNumber {
			// compute precise block time
			blockTime := time.Since(now)
			now = time.Now()

			block, err := c.GetBlockByNumber(ethgo.BlockNumber(blockNumber), false)
			if err != nil {
				panic(err)
			}
			if block == nil {
				// We're not going to continue past this point if we don't have a block
				continue
			}
			lastBlockNumber = blockNumber

			var blockTimestampDiff time.Duration
			var tps float64

			if prevBlock != nil {
				// compute block time
				blockTimestampDiff = time.Unix(int64(block.Timestamp), 0).Sub(time.Unix(int64(prevBlock.Timestamp), 0))
				// Compute TPS
				tps = float64(len(block.TransactionsHashes)) / float64(blockTimestampDiff.Seconds())
			}

			prevBlock = block

			rootTS := metrics.NewRegistry().RootTagSet()
			if c.vu != nil || c.vu.State() != nil || rootTS != nil {
				if _, loaded := blocks.LoadOrStore(c.opts.URL+strconv.FormatUint(blockNumber, 10), true); loaded {
					// We already have a block number for this client, so we can skip this
					continue
				}

				metrics.PushIfNotDone(c.vu.Context(), c.vu.State().Samples, metrics.ConnectedSamples{
					Samples: []metrics.Sample{
						{
							TimeSeries: metrics.TimeSeries{
								Metric: c.metrics.Block,
								Tags: rootTS.WithTagsFromMap(map[string]string{
									"transactions": strconv.Itoa(len(block.TransactionsHashes)),
									"gas_used":     strconv.Itoa(int(block.GasUsed)),
									"gas_limit":    strconv.Itoa(int(block.GasLimit)),
								}),
							},
							Value: float64(blockNumber),
							Time:  time.Now(),
						},
						{
							TimeSeries: metrics.TimeSeries{
								Metric: c.metrics.GasUsed,
								Tags: rootTS.WithTagsFromMap(map[string]string{
									"block": strconv.Itoa(int(blockNumber)),
								}),
							},
							Value: float64(block.GasUsed),
							Time:  time.Now(),
						},
						{
							TimeSeries: metrics.TimeSeries{
								Metric: c.metrics.TPS,
								Tags:   rootTS,
							},
							Value: tps,
							Time:  time.Now(),
						},
						{
							TimeSeries: metrics.TimeSeries{
								Metric: c.metrics.BlockTime,
								Tags: rootTS.WithTagsFromMap(map[string]string{
									"block_timestamp_diff": blockTimestampDiff.String(),
								}),
							},
							Value: float64(blockTime.Milliseconds()),
							Time:  time.Now(),
						},
					},
				})
			}
		}
	}
}
