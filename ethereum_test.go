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
	c, _ := jsonrpc.NewClient("http://localhost:10002")
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

func Test_SendRawTransaction(t *testing.T) {
	client, err := setupClient()
	if err != nil {
		t.Fatal(err)
	}

	nonce, err := client.GetNonce(client.w.Address().String())
	if err != nil {
		t.Fatal(err)
	}

	// Deploy the contract
	tx, err := client.SendRawTransaction(Transaction{
		To:    "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
		Value: 1000000000000000000,
		Nonce: nonce,
	})
	if err != nil {
		t.Fatal(err)
	}

	receipt, err := client.WaitForTransactionReceipt(tx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(receipt.BlockHash)
}
