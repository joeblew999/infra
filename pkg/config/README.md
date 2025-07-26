# Config

Formalise config here, so we do not need config files.

## Environment detection

At the moment we have a IsProduction function that use the ENV.


## Environments

In Dev environment we use the local ./.dep folder. This is so that the files are all local and do not pollute your OS at all.

In Prod environment, we use the standard OS folders. This is so that when users uninstall the software and then reinstall it, the data is still there, and also so we respect each OS and the place it stores things normally on disk.


## Later things

Eventually we will enable this to be optionally driven by NATS Jetstream, so that the config is in the KV store.

this will probably we when in Production mode.

when we go down this route, we will need to use Protobufs so we support schema evolution.

We will also need a proper way to config the local nats leaf node and upgrade it when the NATS Cluster changes its streams configurations.

We need to really get the AI to think about this fully before we make a mess.

---

expot as JSON so we can show on CLI and in Web GUI, via Web Server.



