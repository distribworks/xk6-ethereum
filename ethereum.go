// xk6 build --with github.com/grafana/xk6-ethereum=.
package ethereum

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/distribworks/xk6-ethereum/contracts"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/contract"
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

// EstimateGas returns the estimated gas for the given transaction.
func (c *Client) EstimateGas(tx Transaction) (uint64, error) {
	to := ethgo.HexToAddress(tx.To)

	msg := &ethgo.CallMsg{
		From:     c.w.Address(),
		To:       &to,
		Value:    nil,
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
func (c *Client) WaitForTransactionReceipt(hash string) (*Receipt, error) {
	for {
		receipt, err := c.GetTransactionReceipt(hash)
		if err != nil {
			if err.Error() != "not found" {
				return nil, err
			}
		}
		if receipt != nil {
			return receipt, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// DeployLoadTester deploys the load tester contract.
func (c *Client) DeployLoadTester() (string, error) {
	opts := []contract.ContractOption{
		contract.WithJsonRPC(c.client.Eth()),
		contract.WithSender(c.w),
	}

	// deploy the contract
	txn, err := contracts.DeployLoadTester(c.client, c.w.Address(), []interface{}{}, opts...)
	if err != nil {
		return "", err
	}

	err = txn.Do()
	if err != nil {
		return "", err
	}

	receipt, err := txn.Wait()
	if err != nil {
		return "", err
	}

	return receipt.ContractAddress.String(), nil
}

// CallLoadTester calls as specific function of the load tester contract.
func (c *Client) CallLoadTester(contractAddress string, function string, args ...interface{}) (string, error) {
	opts := []contract.ContractOption{
		contract.WithJsonRPC(c.client.Eth()),
		contract.WithSender(c.w),
	}

	// deploy the contract
	lt := contracts.NewLoadTester(ethgo.HexToAddress(contractAddress), opts...)

	var txn contract.Txn
	var err error
	switch function {
	case "inc":
		txn, err = lt.Inc()
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unknown function: %s", function)
	}

	err = txn.Do()
	if err != nil {
		return "", err
	}

	receipt, err := txn.Wait()
	if err != nil {
		return "", err
	}

	return receipt.TransactionHash.String(), nil
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
