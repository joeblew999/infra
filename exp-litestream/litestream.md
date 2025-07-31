# sqlite and ducklake

THIS IS JUST TOP EXPERIMENT.

pkg/dep, now included litstream.

litestream can run a SQLITE and stream bacups to S3, but also make read replicas: https://fly.io/blog/litestream-revamped/#lightweight-read-replicas

https://litestream.io
https://github.com/benbjohnson/litestream/releases/tag/v0.5.0-beta1

go install github.com/benbjohnson/litestream/cmd/litestream@latest

ducklake also uses S3 and Sqlite.

so we can use litestream for BOTH !!

this gets use very close to stateless.

## config

https://litestream.io/reference/config/

Can also replicate to a FS

Litestream can monitor one or more database files that are specified in the configuration file

## pocketbase

Use this, becaause we get all the good stuff.


