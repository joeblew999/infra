# todo

## Litestream 

needs to be compiled from source, so we get version for all Platforms. Still do not know if it will run on windows. I do not see any trikcy CGO, so might work.

- github.com/mattn/go-sqlite3 is CGO 

I might need to just fork Litestream, and repalce the CGO SQL with non CGO. It will make things much easier, despite some slow down. 

Or do my own buidls with Zig ?

The make file sshows the "fly mcp * " commands. This looks like Fly has build in MCP ?? 

https://fly.io/docs/flyctl/mcp/

https://fly.io/docs/mcp/

DEF lets see what we can do with fly MCP !!

Litestream has NATS integrations now. try it

Litestream has Read replicas now, so with NATS we can replciate Master to many Slaves  ? 

