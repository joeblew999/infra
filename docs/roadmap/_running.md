# Development Notes

just for me.

!! AI should not touch this !!

---

deps

what is devs have deps installed on path ? we are fine, because we ONLY use .deps folder fully.

---

caddy

do we need a caddyfile in root for running our self ?

will the docker running on fly be ok with pulling down the caddy binary and running it ? i dont know. I know dockers are meant to be immutable, but i am fine with it.

so we deploy to fly and have many things beind it.
- infra can then pull in other binaries, and run them
- can get fancy later and have those other binaries tell nats, and then have caddyfile configure to the port they are running on.

can use ko ? 

need pkg/caddy ?

can amp cloudflare through

can use nats wrapper so we can see everything from my laptop ?
- the runner will need to log to slog nats in that case.




---

pkg/mcp needs to write to the claude and gemini file, so we can setup MCP servers automatically. 

- Ask AI for help.
- gemimi cli is always getting "cant do strong replace" errors

https://github.com/modelcontextprotocol/registry
https://github.com/ravitemer/mcp-registry types might be cool for us to use ?
Is there officla one ?

---

claude connectors

https://support.anthropic.com/en/articles/11175166-getting-started-with-custom-connectors-using-remote-mcp

exisitng
https://support.anthropic.com/en/articles/11176164-pre-built-integrations-using-remote-mcp


---

tofu to deplyo to fly.io.

after a deploy, we want out .deploy folder to update, so we always have a record of where stuff is running.

we will also need cloudlfare tofu to configure one of our domaains.

we should also make caddy work, so then we can run many Domains --> Caddy --> the infra setup with different data driving it. 

---

pkg/log neds to use https://github.com/samber/slog-nats
pkg/config to model options, so that at startup we can put the system into logging to std out, file, or nats. 

then a page to web so we can see the logs of the system itself ! its self similar.

---

pkg/gops should change to metrics ? 

add health endpoint to web server ? so that any other system can see it

add health endpoint to nats, so any nats system can use it.  

add health page to web sever, so we can see the health data via nats and datastar. self simialr :) 

---

pkg/cmd.go is too big too ? ask gemini for advice.  

---

web/server.go is too big too ? ask gemini for advice. 

---

Need a good MCP to Help gemini to write and debig golang.  See MCP docs.

---


We need an easy way for nodejs, deno and bun devs be able to run my golang binary. 

I have a prototype in pack foolder.   ask gemini for advice on it, but do not change anything yet. Claude made this for us.

---

DatastarUI: https://github.com/CoreyCole/datastarui

- start using it for our Web code.
- it has CLAUDE.md: https://github.com/CoreyCole/datastarui/blob/main/CLAUDE.md
- workout best way for us to use it. The Docker has some aspects: https://github.com/CoreyCole/datastarui/blob/main/docker-compose.yml
- Playwright is controlled off this: https://github.com/CoreyCole/datastarui/blob/main/package.json
  - Maybe we need same in taskfile for Web Server fodler ? 
- add its Playwright MCP, so AI can help me. 

---

Boostrap phases and architypes

phases need to be idempotent, and look for a thing fist being there before moving on. so we have a unidirectional phasing

Architypes is more about the role that we are runnign Infra as. 

- cmd to only boot nats server ? 
  - Because we are dependent on NATS for logging and so many other things, and we want the AI to also start to use it to see everything.

