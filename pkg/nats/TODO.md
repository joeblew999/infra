# todo


Use as a docker container registry.

ko create the OCI and tar 

nats s3 with some wrapper code is a poor many docker registry.


## DEV and PROD and CLUSTER


We do not want too many options. We want idempotency and IAC.

Use the pkg/config. Use .env for overrides.

A Server always boots the NATS Leaf embedded, because this is its offline Server.

A Server always tries to connect to the NATS Cluster, but it might not be up during initiall development. 


NATS LEAF starts first before pretyt much anything else, because we use it for so much.

I do not know if we need health checks for the embedded server yet.

We already have data persistence, and use a volume for the .data folder

## Fly.io NATS Cluster

We will need to run 6 NATS servers in a cluster, one in each region. We will also need a way to do rolling upgrades without downtime. Fly.io can help us here.

https://github.com/fly-apps/nats-cluster is pretty old. 
https://github.com/jeffh/nats-cluster seems to be the latest fork that has what we need. 

## Web GUI

We need a simple web GUI for NATS server to help with configuration and monitoring.

NATS JSON schemas have everything we need to build a web GUI with Datastar and DatastarUI.

We get canonical data with zero translation layer.

Grab the schemas, spin up the SSE bridge, and let Datastar consume the JSON exactly as NATS defines it.

We can safely model the Datastar frontend against these schemas - they are versioned, stable, and guaranteed to match the JSON you receive when you subscribe to $SYS.REQ.* and $JS.EVENT.*.

Check that they are versioned via the nats cli.

The schemas are in https://github.com/nats-io/jsm.go.

```sh
nats schema list          # list all known types  
nats schema show io.nats.jetstream.api.v1.stream_info_response
```


| Kind                               | Example type                                     | Schema URI                                      |
| ---------------------------------- | ------------------------------------------------ | ----------------------------------------------- |
| JetStream API requests & responses | `io.nats.jetstream.api.v1.stream_create_request` | `…/jetstream/api/v1/stream_create_request.json` |
| Server status (VARZ)               | `io.nats.server.v1.varz_response`                | `…/server/v1/varz_response.json`                |
| Connection lists (CONNZ)           | `io.nats.server.v1.connz_response`               | `…/server/v1/connz_response.json`               |
| JetStream advisories               | `io.nats.jetstream.advisory.v1.api_audit`        | `…/jetstream/advisory/v1/api_audit.json`        |




## Other ideas

Can maybe use PocketBase for storing NATS credentials.

Interesting refs:

https://github.com/skeeeon has really good NATS and PocketBase stuff.

https://github.com/skeeeon/pb-nats can mint NATS JWTs and store them in PB, which is highly useful for making new users at runtime.

https://github.com/skeeeon/pb-cli to interact with PB. Zero deps.
- Has a context system, just like NATS, but not using it.
- So designed for multi project use.
- NOT tagged yet.

https://github.com/skeeeon/onvif-nats-gateway for security cameras

