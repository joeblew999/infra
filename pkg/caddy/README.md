# caddy

For Reverve Proxy of things running.

- Bento

- Other ?

Our Services system still has to start the things up of course.

Caddy Config file should be in .data ? 

THis will mean that Caddy is doing the SSL Termination ? or can we still let Cloudflare do the Termination ? 

## Runny caddy locally

The advantgae of using Caddy for SSL Termiantion, is that we can locally run under HTTPS, which we need for certain things we rela on to work in the Browser.

We want Caddy to do the HTTPS when locally.

```sh

# install caddy locally
caddy version   # verify

# quick test
DOMAIN=localhost PORT=443 caddy run --config Caddyfile
# browse to https://localhost  (Caddy issues its own cert)
```

## running caddy on fly.io

Click the link to view conversation with Kimi AI Assistant https://www.kimi.com/share/d2cn4l0c86sc1vsvl050

We can disable Caddy, so that Fly default HTTPS works. 


If we let Caddy do the SSL Termiantion, then the Certs that it generates with Lets Encrypt will need to be stored in NATS, so that when we have many Servers running on Fly.io, its gets them from NATS, and does not ask Lets Enctypt.

It will also mean tha the Fly config probably needs to change.

## Fly.io run (HTTPS on your Cloudflare domain)

```sh

fly launch --no-deploy --name my-app
fly deploy

fly certs create www.example.com          # Fly issues Let’s Encrypt cert
fly certs create example.com              # add the apex too if you want

```

Then on CF

| Type  | Name | Target         | Proxy status |                             |
| ----- | ---- | -------------- | ------------ | --------------------------- |
| CNAME | @    | my-app.fly.dev | DNS-only ⛅   | ← lets Fly verify ownership |
| CNAME | @    | my-app.fly.dev | Proxied 🟠   | ← production traffic        |



Tell Caddy which domain to answer

```sh
flyctl secrets set DOMAIN=www.example.com
flyctl deploy
```

Force HTTPS redirect (optional, done in Cloudflare)
Cloudflare → SSL/TLS → Edge Certificates → “Always Use HTTPS” = ON.

Quick checklis

| Environment | TLS terminates at | Caddy auto\_https | External URL              |
| ----------- | ----------------- | ----------------- | ------------------------- |
| Local       | Caddy itself      | ON (localhost)    | <https://localhost>       |
| Fly.io      | Fly edge proxy    | OFF               | <https://www.example.com> |




