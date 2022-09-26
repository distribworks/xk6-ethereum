// xk6 build --with github.com/grafana/xk6-ethereum=.
package ethereum

import (
	"fmt"
	"math/big"
	"time"

	"github.com/dop251/goja"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
)

type Client struct {
	w       *wallet.Key
	client  *jsonrpc.Client
	chainID *big.Int
	vu      modules.VU
	metrics ethMetrics
}

func (c *Client) Exports() modules.Exports {
	return modules.Exports{}
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
func (c *Client) GetBlockByNumber(number ethgo.BlockNumber) (*ethgo.Block, error) {
	return c.client.Eth().GetBlockByNumber(number, true)
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
	if tx.GasPrice == 0 {
		tx.GasPrice = 5242880
	}

	t := &ethgo.Transaction{
		From:     ethgo.HexToAddress(tx.From),
		To:       &to,
		Value:    big.NewInt(tx.Value),
		Gas:      tx.Gas,
		GasPrice: tx.GasPrice,
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
		From:     c.w.Address(),
		To:       &to,
		Value:    big.NewInt(tx.Value),
		Gas:      gas,
		GasPrice: tx.GasPrice,
		Nonce:    tx.Nonce,
		Input:    tx.Input,
		Type:     ethgo.TransactionLegacy,
		ChainID:  c.chainID,
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
func (c *Client) GetTransactionReceipt(hash string) (*Receipt, error) {
	r, err := c.client.Eth().GetTransactionReceipt(ethgo.HexToHash(hash))
	if err != nil {
		return nil, err
	}

	if r != nil {
		return &Receipt{
			TransactionHash:   r.TransactionHash.String(),
			TransactionIndex:  r.TransactionIndex,
			ContractAddress:   r.ContractAddress.String(),
			BlockHash:         r.BlockHash.String(),
			From:              r.From.String(),
			BlockNumber:       r.BlockNumber,
			GasUsed:           r.GasUsed,
			CumulativeGasUsed: r.CumulativeGasUsed,
			LogsBloom:         r.LogsBloom,
			Status:            r.Status,
		}, nil
	}

	return nil, fmt.Errorf("not found")
}

// WaitForTransactionReceipt waits for the transaction receipt for the given transaction hash.
func (c *Client) WaitForTransactionReceipt(hash string) *goja.Promise {
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
					metrics.PushIfNotDone(c.vu.Context(), c.vu.State().Samples, metrics.ConnectedSamples{
						Samples: []metrics.Sample{
							{
								Metric: c.metrics.TimeToMine,
								Tags:   metrics.NewSampleTags(map[string]string{"call": "get_transaction_receipt"}),
								Value:  float64(time.Since(now) / time.Millisecond),
								Time:   now,
							},
						},
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

// makeHandledPromise will create a promise and return its resolve and reject methods,
// wrapped in such a way that it will block the eventloop from exiting before they are
// called even if the promise isn't resolved by the time the current script ends executing.
func (c *Client) makeHandledPromise() (*goja.Promise, func(interface{}), func(interface{})) {
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
