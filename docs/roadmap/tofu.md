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


## Core Providers

- fly.io to run this server: https://registry.terraform.io/providers/fly-apps/fly/latest/docs

- cloudflare for domains and tunnels: https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs 

- incus: https://registry.terraform.io/providers/lxc/incus/latest/docs

- null: https://registry.terraform.io/providers/hashicorp/null/latest/docs







