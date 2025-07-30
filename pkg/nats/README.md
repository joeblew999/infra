# nats

Nats Jetstreams.

We can run it as a Server or as a Leaf node.

Interesting Refs:

https://github.com/skeeeon has really good NATS and Pocketbase stuff.


https://github.com/skeeeon/pb-nats can mint NATS JWTS, and stroe them in PB, which is highly useful for making new users at runtime.

https://github.com/skeeeon/pb-cli to interact with PB. zero deps.
- Has a context system, just like NATS, but not using it. 
- So designed for multi project use.
- NOT tagged yet.

## AUth and Authz

NATS Auth callout pattern

## Idemptonecy patterns

There is a way to use JSON Schema with NATS, such that when you configure NATS via the Nats cli, that it puts JSON Schema to disk. 

Its like how with a DB, you have DB setup scripts, so that you can sure everything works properly.

