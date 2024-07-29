import eth from 'k6/x/ethereum';

let rpc_url = __ENV.RCP_URL
if (rpc_url == undefined) {
  rpc_url = "http://localhost:8545"
}

// You can use an existing premined account
const root_address = "0x67b1d87101671b127f5f8714789C7192f7ad340e"
let nonce = 0;

export default function (data) {
  const client = new eth.Client({
    url: rpc_url,
    // You can also specify a private key here
    privateKey: '26e86e45f6fc45ec6e2ecd128cec80fa1d1505e5507dcd2ae58c3130a7a97b48',
    // or a mnemonic
    // mnemonic: 'my mnemonic'
  });

  let prev_nonce = client.getNonce(root_address);
  if (nonce < prev_nonce) {
    nonce = prev_nonce;
  }

  console.log(`nonce => ${nonce}`);
  const gas = client.gasPrice();
  console.log(`gas price => ${gas}`);

  const bal = client.getBalance(root_address, client.blockNumber());
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
