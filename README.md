# xk6-ethereum

A k6 extension to interact with EVM based blockchains.

## Getting started

1. [Build](#build) or Install [BlockSpeed](https://github.com/distribworks/blockspeed)

2. Check the examples folder to learn how to use it

## Build

To build a `k6` binary with this plugin, first ensure you have the prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- If you're using SQLite, a build toolchain for your system that includes `gcc` or
  another C compiler. On Debian and derivatives install the `build-essential`
  package. On Windows you can use [tdm-gcc](https://jmeubank.github.io/tdm-gcc/).
  Make sure that `gcc` is in your `PATH`.
- Git

Then:

1. Install `xk6`:
  ```shell
  go install go.k6.io/xk6/cmd/xk6@latest
  ```

2. Build the binary:
  ```shell
  xk6 build --with github.com/distribworks/xk6-ethereum
  ```

## Javascript API

### Module `k6/x/ethereum`

The `k6/x/ethereum` module contains the Ethereum extension to interact with Ethereum RPC API. To import the module add

```javascript
import eth from 'k6/x/ethereum';
```

### Class `eth.Client({[url, mnemonic, privateKey]})`

The class Client is an Ethereum RPC client that can perform several operations to an Ethereum node. The constructor takes the following arguments:

#### Example:
```javascript
import eth from 'k6/x/ethereum';
const client = new eth.Client({
    url: 'http://localhost:8545',
});
```

### Methods 

  - `gasPrice() number`
  - `getBalance(address: string, blockNumber: number) number`
  - `blockNumber() number`
  - `getBlockByNumber(block: number, full: boolean) Block`
  - `getNonce(address: string) number`
  - `estimateGas(tx: Transaction) number`
  - `sendTransaction(tx: Transaction) string`
  - `sendRawTransaction(tx: Transaction) string`
  - `getTransactionReceipt(tx_hash: string) Receipt`
  - `waitForTransactionReceipt(tx_hash: string) => Promise<Receipt>`
  - `accounts() string[]`
  - `newContract(address: string, abi: string) Contract`
  - `deployContract(abi: string, bytecode: string, args[]) Receipt`

### Objects

```
Transaction
{
  from:     string
  to:       string
  input:    object
  gas_price: number
  gas:      number
  value:    number
  nonce:    number
  // eip-2930 values
  chain_id: number
}
```

```
Receipt
{
  transaction_hash:    object
  transaction_index:   number
  contract_address:    string
  block_hash:          object
  from:                string
  block_number:        number
  gas_used:            number
  cumulative_gas_used: number
  logs_bloom:          object
  logs:                Log[]
  status:              number
}
```

```
Log
{
  removed:           bool
  log_index:         number
  transaction_index: number
  transaction_hash:  object
  block_hash:        object
  block_number:      number
  address:           string
  topics:            object[]
  data:              object
}
```

```
Contract{}

txn() Receipt
call() object
```


### Metrics

It exposes the following metrics:

  * ethereum_block: Blocks in the chain during the test
  * ethereum_req_duration: Time taken to perform an API call to the client
  * ethereum_tps: Computation of Transactions Per Second mined
  * ethereum_time_to_mine: Time it took since a transaction was sent to the client and it has been included in a block

### Example

```javascript
import eth from 'k6/x/ethereum';

const client = new eth.Client({
    url: 'http://localhost:8545',
    // You can also specify a private key here
    // privateKey: 'private key of your account',
    // or a mnemonic
    // mnemonic: 'my mnemonic'
});

// You can use an existing premined account
const root_address = "0x85da99c8a7c2c95964c8efd687e95e632fc533d6"

export function setup() {
  return { nonce: client.getNonce(root_address) };
}

export default function (data) {
  console.log(`nonce => ${data.nonce}`);
  const gas = client.gasPrice();
  console.log(`gas price => ${gas}`);

  const bal = client.getBalance(root_address, client.blockNumber());
  console.log(`bal => ${bal}`);
  
  const tx = {
    to: "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
    value: Number(0.0001 * 1e18),
    gas_price: gas,
    nonce: data.nonce,
  };
  
  const txh = client.sendRawTransaction(tx)
  console.log("tx hash => " + txh);
  // Optional: wait for the transaction to be mined
  // const receipt = client.waitForTransactionReceipt(txh).then((receipt) => {
  //   console.log("tx block hash => " + receipt.block_hash);
  //   console.log(typeof receipt.block_number);
  // });
  data.nonce = data.nonce + 1;
}
```
