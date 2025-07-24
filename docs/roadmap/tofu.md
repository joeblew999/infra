# Tofu

We use tofu, instead of Terraform.

Dep system ensures we have it and the right version.

Configs are held in the terraform folder.

## STATUS

pkg/dep ensures tofu binary is download.
pkg/tofu in place to run the tofu binary.
pkg/cmd is hooked up to pkg/tofu, so we can run tofu commands.



## Providers

- fly to run this server.

- cloudflare for domains and tunnels.





