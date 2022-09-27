package ethereum

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	_, err = client.GasPrice()
	require.NoError(t, err)

	// // Deploy the contract
	// _, err = client.EstimateGas(Transaction{
	// 	// To:       "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
	// 	Value:    0,
	// 	Input:    contracts.LoadTesterBin(),
	// 	GasPrice: gas,
	// })
	// if err != nil {
	// 	t.Fatal(err)
	// }
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

	prom := client.WaitForTransactionReceipt(tx)
	t.Log(prom)
}

func Test_DeployContract(t *testing.T) {
	client, err := setupClient()
	require.NoError(t, err)

	contractBin := "6080604052348015600f57600080fd5b50609b8061001e6000396000f3fe6080604052348015600f5" +
		"7600080fd5b506004361060285760003560e01c8063aeecb68814602d575b600080fd5b6033603556" +
		"5b005b60008090505b600a43028110156062576000808154809291906001019190505550808060010" +
		"1915050603b565b5056fea2646970667358221220b69191cdd18045d942bd33c23c74ed334b2604f5" +
		"8490458cc8581e54f8cffed664736f6c63430006060033"

	contractABI := `[{
"burnBlockNumberDependentGas": {
	"inputs": [],
	"name": "burnBlockNumberDependentGas",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function",
  }
}]`

	res, err := client.DeployContract(contractABI, contractBin)
	require.NoError(t, err)
	t.Log(res)
}
