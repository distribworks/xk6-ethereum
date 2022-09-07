import eth from 'k6/x/ethereum';

const client = eth.newClient({
    url: 'http://localhost:8541',
    chainID: 1256
    // You can also specify a private key here
    // privateKey: '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef'
    // or a mnemonic
    // mnemonic: 'my mnemonic'
});

var nonce = client.getNonce("0x85da99c8a7c2c95964c8efd687e95e632fc533d6");

export function setup() {
  const lta = client.deployLoadTester();
  console.log("Load tester deployed at: " + lta);

  return { lta: lta };
}

// Increment the nonce as we've deployed the load tester contract
nonce = nonce + 1;

export default function (data) {
  const gas = client.gasPrice();
  console.log(`gas => ${gas}`);
  
  const block = client.getBlockByNumber(0);
  
  const bal = client.getBalance("0x85da99c8a7c2c95964c8efd687e95e632fc533d6", block.number);
  console.log(`bal => ${bal}`);
  
  const tx = {
    to: "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
    value: 0x38d7ea4c68000,
    gas_price: gas,
    nonce: nonce,
  };
  
  const txh = client.sendTransaction(tx)
  console.log("tx hash => " + txh);
  const receipt = client.waitForTransactionReceipt(txh)
  console.log("tx block hash => " + receipt.block_hash);
  nonce = nonce + 1;

  const f = client.callLoadTester(data.lta, "inc")
  nonce = nonce + 1;
  console.log("call inc => " + f);
}
