# litestream

https://github.com/benbjohnson/litestream

no CGO
has VFS
has real time replcation

We can import it like we do for pocketbase etc so we have full control.

--

example.gox was done by the AI to show using it.

export EXAMPLE_DB_DIR=pb_data
export EXAMPLE_DB_NAME=data.db
export EXAMPLE_S3_URL="s3://your-bucket/pb"
export LITESTREAM_ACCESS_KEY_ID=...     # or use AWS_* envs if configured globally
export LITESTREAM_SECRET_ACCESS_KEY=...
export EXAMPLE_NATS_URL="nats://127.0.0.1:4222"
