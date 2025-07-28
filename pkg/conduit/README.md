# conduit

This is a more complex pkg, but we use the exisitng pkg to help use.

All their repos: https://github.com/orgs/ConduitIO/repositories

Example to try to make sure of stuff works: https://github.com/ConduitIO/conduit-processor-example has example using stand alone processes, which i think is best for failure protection, and easier development.

We need:

- https://github.com/ConduitIO/conduit
- https://github.com/ConduitIO/conduit/releases/tag/v0.14.0

---

- https://github.com/ConduitIO/conduit-connector-s3
- https://github.com/ConduitIO/conduit-connector-s3/releases/tag/v0.9.3

---

- https://github.com/ConduitIO/conduit-connector-postgres
- https://github.com/ConduitIO/conduit-connector-postgres/releases/tag/v0.14.0

## Lab connectors

We will later pull in connectors at https://github.com/orgs/conduitio-labs/repositories

- https://github.com/conduitio-labs/conduit-connector-cosmos-nosql
- Does not have a tag yet

---

- https://github.com/conduitio-labs/conduit-connector-snowflake
- https://github.com/conduitio-labs/conduit-connector-snowflake/releases/tag/v0.3.2

---

- https://github.com/conduitio-labs/conduit-connector-db2
- no tags yet


## Code

Use the pkg/dep and its json format. 




## TODO

1. Need a JSON file like in pkg/dep that we can use to download all the bianries we want.

- What we want it currently in this README, and evolving.

2. Make a dep.go that use pkg/dep and the dep.json to download the binaries. 

## bento

Later, will use https://github.com/warpstreamlabs/bento to help run it.

They have a conduit processor in the works at https://github.com/warpstreamlabs/bento/discussions/396

Code is at https://github.com/gregfurman/bento/tree/add/conduit/internal/impl/conduit for now, until its merged intp main.

This will make running it with bento linking it up to other things work well.

## Testing

Run the tests to verify the package works correctly:

```bash
go test ./pkg/conduit -v
```

For more detailed output:

```bash
go test ./pkg/conduit -v -count=1
```

To test the actual binary creation in the .dep directory:

```bash
# Run tests to create binaries
go test ./pkg/conduit -run TestPackageIntegration -v

# Verify binaries are created
ls -la .dep/conduit*
```

Or use the package directly:

```bash
# Run a quick test
go run -e 'package main; import "github.com/joeblew999/infra/pkg/conduit"; import "log"; func main() { if err := conduit.Ensure(true); err != nil { log.Fatal(err) } }'
```

## Usage

The package provides functions to manage Conduit binaries:

```go
// Ensure all binaries are downloaded
err := conduit.Ensure(false)

// Get path to a specific binary
path := conduit.Get("conduit")
```


