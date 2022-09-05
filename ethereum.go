// xk6 build --with github.com/grafana/xk6-ethereum=.
package ethereum

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
	"go.k6.io/k6/js/modules"
)

const (
	privateKey = "42b6e34dc21598a807dc19d7784c71b2a7a01f6480dc6f58258f78e539f1a1fa"
)

func init() {
	eth := Eth{}

	modules.Register("k6/x/ethereum", &eth)
}

type Eth struct{}

type Client struct {
	w       *wallet.Key
	client  *jsonrpc.Client
	chainID *big.Int
}

// Options defines configuration options for the client.
type Options struct {
	URL        string
	Mnemonic   string
	PrivateKey string
}

type Transaction struct {
	From     string
	To       string
	Input    []byte
	GasPrice uint64
	Gas      uint64
	Value    int64
	Nonce    uint64
	// eip-2930 values
	ChainId int64
}

func (e *Eth) NewClient(options Options) (*Client, error) {
	if options.URL == "" {
		options.URL = "http://localhost:8545"
	}

	if options.PrivateKey == "" {
		options.PrivateKey = privateKey
	}

	var wa *wallet.Key
	if options.Mnemonic != "" {
		w, err := wallet.NewWalletFromMnemonic(options.Mnemonic)
		if err != nil {
			return nil, err
		}
		wa = w
	} else if options.PrivateKey != "" {
		pk, err := hex.DecodeString(options.PrivateKey)
		if err != nil {
			return nil, err
		}
		w, err := wallet.NewWalletFromPrivKey(pk)
		if err != nil {
			return nil, err
		}
		wa = w
	}

	c, err := jsonrpc.NewClient(options.URL)
	if err != nil {
		return nil, err
	}

	cid, err := c.Eth().ChainID()
	if err != nil {
		return nil, err
	}

	return &Client{client: c, w: wa, chainID: cid}, nil
}

func (c *Client) GasPrice() (uint64, error) {
	return c.client.Eth().GasPrice()
}

func (c *Client) GetBalance(address string, blockNumber ethgo.BlockNumber) (uint64, error) {
	b, err := c.client.Eth().GetBalance(ethgo.HexToAddress(address), blockNumber)
	return b.Uint64(), err
}

// GetBlockByNumber returns the block with the given block number.
func (c *Client) GetBlockByNumber(number ethgo.BlockNumber) (*ethgo.Block, error) {
	return c.client.Eth().GetBlockByNumber(number, true)
}

// GetNonce returns the nonce for the given address.
func (c *Client) GetNonce(address string) (uint64, error) {
	return c.client.Eth().GetNonce(ethgo.HexToAddress(address), ethgo.Latest)
}

// SendTransaction signs and sends transaction to the network.
func (c *Client) SendTransaction(tx Transaction) (string, error) {
	to := ethgo.HexToAddress(tx.To)

	gas, err := c.client.Eth().EstimateGas(&ethgo.CallMsg{
		From:     c.w.Address(),
		To:       &to,
		Value:    big.NewInt(tx.Value),
		Data:     tx.Input,
		GasPrice: tx.GasPrice,
	})
	if err != nil {
		return "", fmt.Errorf("failed to estimate gas: %e", err)
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

	t.Gas = gas

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
