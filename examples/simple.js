import eth from 'k6/x/ethereum';
import "@ethersproject/shims";
import * as ethers from "ethers";

let rpc_url = __ENV.RCP_URL
if (rpc_url == undefined) {
  rpc_url = "http://localhost:10002"
}

// You can use an existing premined account
const root_address = "0x85da99c8a7c2c95964c8efd687e95e632fc533d6"
let nonce = 0;

export default function (data) {
  const client = new eth.Client({
    url: rpc_url,
    // You can also specify a private key here
    // privateKey: 'private key of your account',
    // or a mnemonic
    // mnemonic: 'my mnemonic'
  });

  // To connect to a custom URL:
  let client2 = new ethers.providers.JsonRpcProvider(rpc_url);

  client2.getBlockNumber().then((blockNumber) => {
    console.log("blockNumber => " + blockNumber);
  });

  let prev_nonce = client.getNonce(root_address);
  if (nonce < prev_nonce) {1
    nonce = prev_nonce;
  }

  console.log(`nonce => ${nonce}`);
  const gas = client.gasPrice();
  console.log(`gas price => ${gas}`);

  // const bal = client.getBalance(root_address, client.blockNumber());
  const bal = ethers.utils.formatEther(balance);
  console.log(`bal => ${bal}`);
  
  const tx = {
    to: "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
    value: Number(0.0001 * 1e18),
    gas_price: gas,
    nonce: nonce,
  };
  
  const txh = client.sendRawTransaction(tx)
  console.log("tx hash => " + txh);
  // Optional: wait for the transaction to be mined
  // const receipt = client.waitForTransactionReceipt(txh).then((receipt) => {
  //   console.log("tx block hash => " + receipt.block_hash);
  //   console.log(typeof receipt.block_number);
  // });
  nonce++;
}
