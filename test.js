import eth from 'k6/x/ethereum';

const client = eth.newClient({
    url: 'http://localhost:8541',
    chainID: 1256
});

var nonce = client.getNonce("0x85da99c8a7c2c95964c8efd687e95e632fc533d6");

export default function () {
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

  nonce = nonce + 1;
  console.log(client.sendTransaction(tx));
}
