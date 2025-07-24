# Development Notes

just for me.

!! AI should not touch this !!

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

