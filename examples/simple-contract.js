import eth from 'k6/x/ethereum';

const contract_bin = open("./contracts/GasBurner.bin");
const contract_abi = open("./contracts/GasBurner.abi");

// You can use an existing premined account
const root_address = "0x85da99c8a7c2c95964c8efd687e95e632fc533d6"
const url = "http://localhost:10002"
const client = new eth.Client({url: url});

export function setup() {
  const data = {};

  const receipt = client.deployContract(contract_abi, contract_bin)
  
  data.contract_address = receipt.contract_address;
  data.nonce = client.getNonce(root_address);
  
  return data;
}

var nonce = 0;

// VU client
export default function (data) {
  console.log(JSON.stringify(data));

  const con = client.newContract(data.contract_address, contract_abi);
  const res = con.txn("burnGas", 10);
  console.log(`txn hash => ${res.transaction_hash}`);
  console.log(`gas used => ${res.gas_used}`);
  console.log(JSON.stringify(con.call("getTotal")));
}
