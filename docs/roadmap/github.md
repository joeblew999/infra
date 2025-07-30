# github

Then i really need you to tell me how you can use the Claude MCP server for github.

Cause i want you to be able to interact with this proejects github, so you can see everything like Issues, respond to Issues. Later also to do a PR into Github, and then close an Issues related to it after you have validated that it works.

## nats

https://github.com/nats-io/nsc/releases/tag/v2.11.0 will help with minting nats crednetials. 



https://github.com/nats-io/nats-surveyor will help with seeing whats the NATS system is doing globally.  System accounts must be enabled to use surveyor - i guess so it can see everything. 

nats-surveyor.yaml

```yml
servers: nats://127.0.0.1:4222
accounts: true
log-level: debug
```

Per account monitoirng looks cool: https://github.com/nats-io/nats-surveyor?tab=readme-ov-file#jetstream


is there a MCP for prom ? 




