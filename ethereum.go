// xk6 build --with github.com/grafana/xk6-ethereum=.
package ethereum

import (
	"fmt"
	"math/big"
	"strings"

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

	var wa *wallet.Key
	if options.Mnemonic != "" {
		w, err := wallet.NewWalletFromMnemonic(options.Mnemonic)
		if err != nil {
			return nil, err
		}
		wa = w
	} else if options.PrivateKey != "" {
		w, err := wallet.NewWalletFromPrivKey([]byte(options.PrivateKey))
		if err != nil {
			return nil, err
		}
		wa = w
	} else {
		w, err := wallet.NewWalletFromPrivKey([]byte(privateKey))
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

// newTransaction creates a new ethgo transaction.
func (c *Client) newTransaction(tx Transaction) (*ethgo.Transaction, error) {
	t := &ethgo.Transaction{}
	to := ethgo.HexToAddress(strings.ToLower(tx.To))

	t.To = &to
	t.Value = big.NewInt(tx.Value)
	t.Gas = tx.Gas
	t.GasPrice = tx.GasPrice
	t.Nonce = tx.Nonce
	t.Input = tx.Input
	t.Type = ethgo.TransactionLegacy
	t.ChainID = c.chainID

	return t, nil
}

// SendTransaction signs and sends transaction to the network.
func (c *Client) SendTransaction(tx Transaction) (ethgo.Hash, error) {
	t, err := c.newTransaction(tx)
	if err != nil {
		return ethgo.Hash{}, err
	}

	// gas, err := c.client.Eth().EstimateGas(&ethgo.CallMsg{
	// 	From:     ethgo.HexToAddress(tx.From),
	// 	To:       t.To,
	// 	Value:    t.Value,
	// 	Data:     t.Input,
	// 	GasPrice: t.GasPrice,
	// })
	// if err != nil {
	// 	return ethgo.Hash{}, fmt.Errorf("failed to estimate gas: %e", err)
	// }
	//t.Gas = gas

	s := wallet.NewEIP155Signer(t.ChainID.Uint64())
	st, err := s.SignTx(t, c.w)
	if err != nil {
		return ethgo.Hash{}, err
	}

	h, err := st.GetHash()
	if err != nil {
		return ethgo.Hash{}, err
	}
	st.Hash = h

	fmt.Println("sending tx", st.From.String(), st.To.String(), st.Value.String(), st.Gas, st.GasPrice, st.Nonce, st.ChainID, st.Type)
	trlp, err := st.MarshalRLPTo(nil)
	if err != nil {
		return ethgo.Hash{}, fmt.Errorf("failed to marshal tx: %e", err)
	}
	j, _ := st.MarshalJSON()
	fmt.Println(string(j))

	return c.client.Eth().SendRawTransaction(trlp)
}
