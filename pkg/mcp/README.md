# MCP

We want to use Claude code at dev time and runtime.

Because MCP's are built in to many languages we should use docker. We might change that later, when we can start using golang MCP Servers. 

In the End, this will provide teams to build AI systems with their Products.



CLaude make it all here : https://claude.ai/public/artifacts/63fe381a-3e8d-44c6-a5b6-0371df07f6cc

- use docker nats, pocketbase, datastar 

- integates with Claude Code.

---

https://github.com/skeeeon has really good NATS and Pocketbase stuff.


https://github.com/skeeeon/pb-nats can mint NATS JWTS, and stroe them in PB, which is highly useful for making new users at runtime.

https://github.com/skeeeon/pb-cli to interact with PB. zero deps.
- Has a context system, just like NATS, but not using it. 
- So designed for multi project use.
- NOT tagged yet.

---

https://github.com/benallfree/pocketpages provides a JS based way to use Pocketbase. And has Datastar integrated. 

- docs: https://pocketpages.dev/docs

WE DEF need to try this with bun, as its way easier than npm.

This will be highly useful for Devs and Users to extend the system at dev and runtime.

https://github.com/benallfree/pocketpages/tree/main/packages/plugins/datastar




