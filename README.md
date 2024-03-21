## Usage

### Server side

On a publicly-accessible server of your choice, run the following command:

```bash
dev-webhook-relay server -token <your-secret-token>
```

This will start a server on port `8080` (can be customized using `-port`) that listens for incoming webhooks on the
`/webhook` endpoint (can be customized using `-http-endpoint`).

### Client side

On the client side, run the following command:

```bash
dev-webhook-relay client -server ws://<your-server-address>:8080 -token <your-secret-token> -forward-url <your-forward-url>
```

This will start a client that connects to the server and listens for incoming webhooks. When a webhook is received, it
will be forwarded to the specified `forward-url`. The provided `token` must match the server's token.

If the server is behind a TLS-terminating proxy, use `wss://` instead of `ws://`.