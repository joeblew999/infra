# README



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

  ## Task files

  Can run on any device in CI, CD and Production.

  DRY everywhere.

  Designed for any OS.
  