# Tofu

We use tofu, instead of Terraform.

Dep system ensures we have it and the right version.

Terraform configs are held in the terraform folder

## STATUS

pkg/store knows the default Terraform configs folder.
pkg/dep ensures tofu binary is download.
pkg/tofu in place to run the tofu binary.
pkg/cmd is hooked up to pkg/tofu, so we can run tofu commands.

```sh
./infra --mode cli tofu providers

Providers required by configuration:
.
└── provider[registry.opentofu.org/lxc/incus] ~> 0.1

```


## Providers

- fly to run this server.

- cloudflare for domains and tunnels.





