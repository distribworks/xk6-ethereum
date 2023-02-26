package ethereum

import (
	"fmt"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/contract"
)

// Contract exposes a contract
type Contract struct {
	*contract.Contract
	client *Client
}

// Call executes a call on the contract
func (c *Contract) Call(method string, args ...interface{}) (map[string]interface{}, error) {
	return c.Contract.Call(method, ethgo.Latest, args...)
}

// Txn executes a transactions on the contract and waits for it to be mined
// TODO maybe use promise
func (c *Contract) Txn(method string, opts contract.TxnOpts, args ...interface{}) (string, error) {
	txn, err := c.Contract.Txn(method, args...)
	if err != nil {
		return "", err
	}

	txn.WithOpts(&opts)

	err = txn.Do()
	if err != nil {
		return "", fmt.Errorf("failed to send contract transaction: %w", err)
	}

	return txn.Hash().String(), nil
}
