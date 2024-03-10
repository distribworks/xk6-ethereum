import eth from 'k6/x/ethereum';

const contract = JSON.parse(open("./contracts/PolygonZkEVMBridgeV2.json"));

// You can use an existing premined account
const root_address = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266";
const url = "http://localhost:8123";
const client = new eth.Client({
    privateKey: "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
    url: url
});
const bridge = "0x80a540502706aa690476D5534e26939894559c05";

// VU client
export default function () {
  const con = client.newContract(bridge, JSON.stringify(contract.abi));

  let nonce = client.getNonce(root_address);
  let txhash = con.txn(
    "bridgeAsset", { gas_limit: 100000, nonce: nonce },
    0x0,
    root_address,
    0.0001 * 1e18,
    "0x0000000000000000000000000000000000000000",
    false,
    "",
  );
  
  console.log(`txn hash => ${txhash}`);

  client.waitForTransactionReceipt(txhash).then((receipt) => {
    console.log("tx block hash => " + receipt.block_hash);
  });

  console.log(JSON.stringify(client.getBalance(root_address)));
}
