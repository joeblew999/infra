# README

https://github.com/joeblew999/infra

AI and Task fiels to help with many golang thngs.

## Concept

Taskfiles can be included, just like a golang package, allowing reusing across your projects.

See: https://taskfile.dev/next/experiments/remote-taskfiles/#including-remote-taskfiles

```yaml
version: '3'
includes:
  remote: https://:{{.TOKEN}}@{{.REDACTED_URL}}.git/Taskfile.dist.yml?ref=master
```

Also they can run on your laptop, in CI ( github actions), in CD ( tarraform via taskfile ) and also in Production.

It helps to make things DRY. The most important thing is that the Taskfiles and golang is versioned, so if this repo changes, it will not break your repo using it. Just use git hashs or git tags as you please.

## AI

Is setup for Claude CLI and Gemini CLI.

I find Copilot and the extensions to be really heavy and make VSCOde slow.

Will might add Support for VSCODE Copilot, if requested I find it makes the IDE slow, but https://code.visualstudio.com/mcp support might be good.

If you want another AI setup let me know.

## Deps

task: https://github.com/go-task/task, https://github.com/go-task/task/releases/tag/v3.44.1

caddy: https://github.com/caddyserver/caddy, https://github.com/caddyserver/caddy/releases/tag/v2.10.0

tofu: https://github.com/tofuutils, https://github.com/tofuutils/tenv

bento: https://github.com/warpstreamlabs/bento, https://github.com/warpstreamlabs/bento/releases/tag/v1.9.0

incus: https://github.com/lxc/incus, https://github.com/lxc/incus/releases/tag/v6.14.0

## Deployment

Origin on Hetzner in Germany.
- to Get European Coverage

Secondaries on OVH in for the rest of the world.
- https://www.ovhcloud.com/en-au/
- https://github.com/ovh/terraform-provider-ovh

Supported Resources: The provider supports a wide range of OVH services including:

Public Cloud instances and infrastructure
Dedicated servers
Load balancers
DNS zones and records
vRack networking
Kubernetes clusters
Object storage
Databases

## Stack of bits i am using this for.




 infra

  Hetzner only. Dedicated or VPS. ZFS required.

  Incus: https://github.com/lxc/incus
  Docs: https://linuxcontainers.org/incus/docs/main/
  Install: https://linuxcontainers.org/incus/docs/main/installing/
  Third party: https://linuxcontainers.org/incus/docs/main/third_party/

  Taskfile for darwin, linux, windows. See 'task dep-cli'.

  Remote uses Tofu.
  Terraform: https://github.com/lxc/terraform-provider-incus
  CI: https://github.com/cloudbase/garm

  Hetzner AX42: https://www.hetzner.com/dedicated-rootserver/ax52/
  64GB RAM, 2x1TB SSD, $59/mo

  Reference: https://pieterbakker.com/incus/

  FS is the DB.
  ko builds containers.

  NATS for everything. NGS for tracking. Cloudflare R2 for fast file requests.
  Web/Desktop/Mobile via Web. NATS leaf node/code on desktops.

  This was the actual infrastructure repository README content before it got replaced
  with the Hetzner Cloud CLI documentation.