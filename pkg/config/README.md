# Config

we do not want config files, but instead use this to formalise config here.

Eventually we will enable this to be optionally driven by NATS Jetstream, so that the config is in the KV store.

this will probably we when in Production mode.

when we go down this route, we will need to use Protobufs so we support schema evolution.

We will also need a proper way to config the local nats leaf node and upgrade it when the NATS Cluster changes its streams configurations.

We need to really get the AI to think about this fully before we make a mess.

---

expot as JSON so we can show on CLI and in Web GUI, via Web Server.



