import eth from 'k6/x/ethereum';

const client = eth.newClient({
    url: 'http://localhost:8541',
    chainID: 1256
});

export default function () {
  const gas = client.gasPrice();
  console.log(`gas => ${gas}`);

  const block = client.getBlockByNumber(0);

  const bal = client.getBalance("0x85da99c8a7c2c95964c8efd687e95e632fc533d6", block.number);
  console.log(`bal => ${bal}`);

  const tx = {
    from: "0x85da99c8a7c2c95964c8efd687e95e632fc533d6",
    to: "0xDEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
    value: 0x38d7ea4c68000,
    gas: 0x5208,
    gas_price: gas,
    nonce: 0x45,
    };

  client.sendTransaction(tx);
}
