package ethereum

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/wallet"
)

func Test_newTransaction(t *testing.T) {
	testAdd := ethgo.HexToAddress("0x0000000000000000000000000000000000000000")

	type args struct {
		tx *Transaction
	}
	tests := []struct {
		name string
		args args
		want *ethgo.Transaction
	}{
		{
			"test conversion",
			args{
				&Transaction{
					From:     "0x0000000000000000000000000000000000000000",
					To:       "0x0000000000000000000000000000000000000000",
					Input:    []byte{},
					GasPrice: 0,
					Gas:      0,
					Value:    0,
					Nonce:    0,
				}},
			&ethgo.Transaction{
				From:     ethgo.HexToAddress("0x0000000000000000000000000000000000000000"),
				To:       &testAdd,
				Input:    []byte{},
				GasPrice: 0,
				Gas:      0,
				Value:    big.NewInt(0),
				Nonce:    0,
				Type:     ethgo.TransactionLegacy,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			if got, _ := c.newTransaction(*tt.args.tx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTransaction() = %v, want %v", got, tt.want)

				w, _ := wallet.NewWalletFromPrivKey([]byte(privateKey))
				s := wallet.NewEIP155Signer(1256)
				_, _ = s.SignTx(tt.want, w)
			}
		})
	}
}
