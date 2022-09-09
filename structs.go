package ethereum

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

type Receipt struct {
	TransactionHash   string
	TransactionIndex  uint64
	ContractAddress   string
	BlockHash         string
	From              string
	BlockNumber       uint64
	GasUsed           uint64
	CumulativeGasUsed uint64
	LogsBloom         []byte
	// Logs              []*Log
	Status uint64
}
