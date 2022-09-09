package ethereum

import (
	"encoding/hex"
	"testing"

	"github.com/distribworks/xk6-ethereum/contracts"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
)

func setupClient() (*Client, error) {
	// Create a new client
	pk, _ := hex.DecodeString("42b6e34dc21598a807dc19d7784c71b2a7a01f6480dc6f58258f78e539f1a1fa")
	wa, _ := wallet.NewWalletFromPrivKey(pk)
	c, _ := jsonrpc.NewClient("http://localhost:8545")
	cid, err := c.Eth().ChainID()
	if err != nil {
		return nil, err
	}

	return &Client{
		client:  c,
		w:       wa,
		chainID: cid,
	}, nil
}

func Test_DeployLoadTester(t *testing.T) {
	// Create a new client
	client, err := setupClient()
	if err != nil {
		t.Fatal(err)
	}

	// Deploy the contract
	_, err = client.DeployLoadTester()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_EstimateGas(t *testing.T) {
	client, err := setupClient()
	if err != nil {
		t.Fatal(err)
	}

	gas, err := client.GasPrice()
	if err != nil {
		t.Fatal(err)
	}

	// Deploy the contract
	_, err = client.EstimateGas(Transaction{
		// To:       "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
		Value:    0,
		Input:    contracts.LoadTesterBin(),
		GasPrice: gas,
	})
	if err != nil {
		t.Fatal(err)
	}
}
