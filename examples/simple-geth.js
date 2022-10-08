import eth from 'k6/x/ethereum';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const client = new eth.Client({});
const root_address = "0x85da99c8a7c2c95964c8efd687e95e632fc533d6"

export function setup() {
  const accounts = client.accounts();
  // If there's not accounts we are not running in dev mode
  if (accounts.length != 0) {
    // Transfer some funds from the coinbase address to the test account if needed
    const bal = client.getBalance(root_address, client.blockNumber());
    if (bal < Number(1000 * 1e18)) {
      const txh = client.sendTransaction({
        from: accounts[0],
        to: root_address,
        value: Number(1000 * 1e18),
      });
      const rcp = client.waitForTransactionReceipt(txh)
    }
  }

  const lta = client.deployLoadTester();
  console.log("Load tester deployed at: " + lta);

  return { lta: lta, nonce: client.getNonce(root_address) };
}

export default function (data) {
  console.log(`nonce => ${data.nonce}`);
  const gas = client.gasPrice();
  console.log(`gas => ${gas}`);

  const bal = client.getBalance(root_address, client.blockNumber());
  console.log(`bal => ${bal}`);
  
  // Get a random eth value between 0.001 and 0.002
  const value = Number(randomIntBetween(1, 20) * 1e16);
  const tx = {
    to: "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
    value: value,
    gas_price: gas,
    nonce: data.nonce,
  };
  
  const txh = client.sendRawTransaction(tx)
  console.log("tx hash => " + txh);

  // Optionally wait for the transaction to be mined
  // const receipt = client.waitForTransactionReceipt(txh)
  // console.log("tx block hash => " + receipt.block_hash);
  data.nonce = data.nonce + 1;
}
