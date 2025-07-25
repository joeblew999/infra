# README

https://github.com/joeblew999/infra

# Endponts

http://localhost:1337/
http://localhost:1337/docs/



## Concept

Its Self Similar design. 

At Dev time, you and the AI can easily do complex flows, because we are using binaries and their CLI to loop. We then formalsie common workflows looping through many binaries in the workflows, ensuring idempotency.    

At Runtime, in Dev and Prod, we do the same thing. There is no shifting in thinking or Architetcure.

When things run, they log, but they can log to NATS too, and so we can self reflect on it at runtime from anywhere. 

We can also publish events to nats when thigns happen to help workflows works well.


Also they can run on your laptop, in CI ( github actions), in CD ( tarraform via taskfile ) and also in Production.

It helps to make things DRY. The most important thing is that the Taskfiles and golang is versioned, so if this repo changes, it will not break your repo using it. Just use git hashs or git tags as you please.

## AI

Is setup for Claude CLI and Gemini CLI.

I find Copilot and the extensions to be really heavy and make VSCOde slow.

Will might add Support for VSCODE Copilot, if requested I find it makes the IDE slow, but https://code.visualstudio.com/mcp support might be good.

If you want another AI setup let me know.

## Deps

See ./roadmap/dep.md



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