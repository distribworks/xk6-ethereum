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

const client = eth.Client({
  url: 'http://localhost:10002',
});
const root_address = "0x85da99c8a7c2c95964c8efd687e95e632fc533d6"

export function setup() {
  return { nonce: client.getNonce(root_address) };
}

export default function (data) {
  console.log(`nonce => ${data.nonce}`);
  const gas = client.gasPrice();
  console.log(`gas => ${gas}`);

  const bal = client.getBalance(root_address, client.blockNumber());
  console.log(`bal => ${bal}`);
  
  const tx = {
    to: "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
    value: Number(0.001 * 1e18),
    gas_price: gas,
    nonce: data.nonce,
  };
  
  const txh = client.sendRawTransaction(tx)
  console.log("tx hash => " + txh);
  data.nonce = data.nonce + 1;
}
