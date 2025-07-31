# sqlite and ducklake

docs: https://litestream.io/


config: https://litestream.io/reference/config/


## WHY ? 

litestream can stream backups to S3 and FS

Can make replicas: https://fly.io/blog/litestream-revamped/#lightweight-read-replicas

So we can use it for our Pocketbase SQLITE, and later ducklake

this gets use very close to stateless, because if the SQLITE DB is not there when we boot, it will restore the DB off the S3.

I think locally, we can us the local FS, instead of the S3. Tests will show if it realyl can.

Can also replicate to a FS

Litestream can monitor one or more database files that are specified in the configuration file



