package ethereum

import (
	"fmt"
	"math/big"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/contract"
)

// Contract exposes a contract
type Contract struct {
	*contract.Contract
	client *Client
}

type TxnOpts struct {
	Value    uint64
	GasPrice uint64
	GasLimit uint64
	Nonce    uint64
}

// Call executes a call on the contract
func (c *Contract) Call(method string, args ...interface{}) (map[string]interface{}, error) {
	return c.Contract.Call(method, ethgo.Latest, args...)
}

// Txn executes a transactions on the contract and waits for it to be mined
// TODO maybe use promise
func (c *Contract) Txn(method string, opts TxnOpts, args ...interface{}) (string, error) {
	txn, err := c.Contract.Txn(method, args...)
	if err != nil {
		return "", fmt.Errorf("failed to create contract transaction: %w", err)
	}

	txo := contract.TxnOpts{
		Value:    big.NewInt(int64(opts.Value)),
		GasPrice: opts.GasPrice,
		GasLimit: opts.GasLimit,
		Nonce:    opts.Nonce,
	}
	txn.WithOpts(&txo)

	err = txn.Do()
	if err != nil {
		return "", fmt.Errorf("failed to send contract transaction: %w", err)
	}

	return txn.Hash().String(), nil
}
