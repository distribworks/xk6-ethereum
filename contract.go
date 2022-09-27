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
func (c *Contract) Txn(method string, args ...interface{}) (*ethgo.Receipt, error) {
	txn, err := c.Contract.Txn(method, args...)
	if err != nil {
		return nil, err
	}

	err = txn.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send contract transaction: %w", err)
	}

	receipt, err := txn.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed waiting for contract transaction: %w", err)
	}

	return receipt, nil
}
