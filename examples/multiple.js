import eth from 'k6/x/ethereum';
import exec from 'k6/execution';
import { fundTestAccounts } from '../helpers/init.js';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.2/index.js';

// export const options = {
//   stages: [
//     { duration: '30s', target: 50 },
//     { duration: '15s', target: 25 },
//     { duration: '15s', target: 0 },
//   ],
// };

export const options = {
  setupTimeout: '220s',
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 200,
      timeUnit: '1s',
      duration: '10s',
      preAllocatedVUs: 10,
      maxVUs: 10,
    },
  },
};

const url = "http://localhost:10002"
// You can use an existing premined account
const root_address = "0x85da99c8a7c2c95964c8efd687e95e632fc533d6";

export function setup() {
  return {accounts: fundTestAccounts(root_address, url)};
}

var nonce = 0;

var client;

// VU client
export default function (data) {
  if (client == null) {
    client = new eth.Client({
      url: url,
      privateKey: data.accounts[exec.vu.idInInstance - 1].private_key
    });
  }

  console.log(`nonce => ${nonce}`);
  
  const tx = {
    to: "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
    value: Number(0.0001 * 1e18),
    gas_price: client.gasPrice(),
    nonce: nonce,
  };

  const txh = client.sendRawTransaction(tx);
  console.log("tx hash => " + txh);
  nonce++;

  // client.waitForTransactionReceipt(txh).then((receipt) => {
  //   console.log("tx block hash => " + receipt.block_hash);
  // });
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }), // Show the text summary to stdout...
    'summary.json': JSON.stringify(data),
  };
}
