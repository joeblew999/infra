# auth

A simple AUth system using Passkey ( so now passwords ) and NATS for sessions and users own streams.

servers session are stored in nats KV. 

Can login via browser and then use nats cli to connec tot your streams.

## HTPPS

HTTPS is required for WebAuthn/passkeys to work, so this automatically uses the pkg/caddy to gen a config and run caddy for us.

## Autofill

Yes, so user does not need to rememebr their username, and OS shows nice chrome.

## TODO

Its sot of working. 

- need all variables relating to origins as const or something. When i tried to test it on my IOS device, it failed. We had to do a hack to try, and so we need to sort this out properly.  Maybe Cloudflare tunnel or something, using on of our real domain names ? 





