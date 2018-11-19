# Starlight SDK

This is a demo TypeScript SDK for [Starlight](https://github.com/interstellar/starlight/tree/main/starlight). It allows you to programmatically create and use payment channels on the Stellar testnet.

The documentation is available [here](https://interstellar.github.io/starlight).

To use this SDK, you first need to [install and run](https://github.com/interstellar/starlight/tree/main/starlight#setup) a local instance of `starlightd`.

```sh
$ starlightd
```

Once you have `starlightd` running, you can programmatically connect to it using this SDK:

```ts
import { Client } from 'starlight-sdk'

async function demo() {
  const client = new Client('http://localhost:7000')

  await client.configInit({
    Username: 'alice',
    Password: 'password',
    HorizonURL: 'https://horizon-testnet.stellar.org/',
  })
  client.subscribe(update => {
    console.log('Received update:', update)
  })
}

demo()
```

You can learn more about the methods available on the Client class in the [documentation](https://interstellar.github.io/starlight/Client.html).

For an example of an app built using this SDK, you can see the code for the [Starlight wallet](https://github.com/interstellar/starlight/tree/main/starlight/wallet).
