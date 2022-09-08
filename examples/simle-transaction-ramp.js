import eth from 'k6/x/ethereum';

export const options = {
  scenarios: {
    example_scenario: {
      executor: 'ramping-arrival-rate',

      // common scenario configuration
      startTime: '0s',
      gracefulStop: '5s',

      // executor-specific configuration
      stages: [
        // It should start 300 iterations per `timeUnit` for the first minute.
        { target: 10, duration: '5s' },

        // It should linearly ramp-up to starting 600 iterations per `timeUnit` over the following two minutes.
        { target: 50, duration: '20s' },

        // It should continue starting 100 iterations per `timeUnit` for the following four minutes.
        { target: 100, duration: '50s' },
      ],
      preAllocatedVUs: 1,
    },
  },
};

const client = eth.newClient({});

const root_address = "0x85da99c8a7c2c95964c8efd687e95e632fc533d6"
var nonce = client.getNonce(root_address);

export function setup() {
  const accounts = client.accounts();
  // If there's not accounts we are not running in dev mode
  if (accounts.length != 0) {
    // Transfer some funds from the coinbase address to the test account
    const txh = client.sendTransaction({
      from: accounts[0],
      to: root_address,
      value: 100000000000000000000,
    });
    client.waitForTransactionReceipt(txh)
  }
}

export default function (data) {
  const bal = client.getBalance(root_address, client.blockNumber());
  console.log(`bal => ${bal}`);
  
  const gas = client.gasPrice();
  console.log(`gas => ${gas}`);
  
  const tx = {
    to: "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
    value: 1000000000000000000,
    gas_price: gas,
    nonce: nonce,
  };
  
  const txh = client.sendRawTransaction(tx)
  console.log("tx hash => " + txh);
  
  nonce = nonce + 1;
}
