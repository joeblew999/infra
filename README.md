# infra

SOURCE : https://github.com/joeblew999/infra

Designed for Hetzner for Dedciated and VPS where ZFS can be used for both.

Taskfile for local and remote setup for darwin, linux and windows.

See "task dep-cli" for pattern for how to do it.


## Incus

code: https://github.com/lxc/incus 

docs: https://linuxcontainers.org/incus/docs/main/

https://linuxcontainers.org/incus/docs/main/third_party/

Remote uses Tofu of course.

## Installation

https://linuxcontainers.org/incus/docs/main/installing/


## OPS

terraform: https://github.com/lxc/terraform-provider-incus

CI : https://github.com/cloudbase/garm



Hetzner Servers:

AX Servers.

AX42 : https://www.hetzner.com/dedicated-rootserver/ax52/
- 64 GB RAM
- 1 TB x 2 SSD
- 59 USD / month


https://pieterbakker.com/blog/ uses Incus on Hetzner it alot.

https://pieterbakker.com/incus/

https://pieterbakker.com/rescaling-incus-zfs-storage-on-hetzner-cloud/


Dedeciated or Cloud can run it.

FS is the DB.

---

ko to build containers.


---

NATS for everything, using NGS for tracking ONLY.

NATS Servers.

NATS Leaf Nodes

nats.js for auth

nats.go for auth, and everythign else

---

Cloudflare 

Only R2 so that file requests are faster.

---

Web, Desktop and Mobile using only Web.

Nats Leaf node and code on desktops.


