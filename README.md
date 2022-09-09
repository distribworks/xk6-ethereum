# xk6-ethereum

A k6 extension to interact with EVM based blockchains.

## Getting started

1. Install `blockspeed`

2. Check the examples folder to learn how to use it

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

### Metrics

#### Request metrics

#### TimeToMine

### Example

```javascript
import eth from 'k6/x/ethereum';
import { utils } from "https://cdn.ethers.io/lib/ethers-5.6.umd.min.js"

const client = new eth.Client({
    url: 'http://localhost:8541',
});

// You can use an existing account with funds
const root_address = "0x85da99c8a7c2c95964c8efd687e95e632fc533d6"
var nonce = client.getNonce(root_address);

export default function (data) {
  const gas = client.gasPrice();
  console.log(`gas => ${gas}`);

  const bal = client.getBalance(root_address, client.blockNumber());
  console.log(`bal => ${bal}`);
  
  const tx = {
    to: "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
    value: utils.parseEther("0.001"),
    gas_price: gas,
    nonce: nonce,
  };
  
  const txh = client.sendRawTransaction(tx)
  console.log("tx hash => " + txh);
  nonce = nonce + 1;
}
```
