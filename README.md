# infra

SOURCE : https://github.com/joeblew999/infra


Origin is ONLY Hetzner.

Incus

code: https://github.com/lxc/incus on Hetzner which has zfs backed in.

docs: https://linuxcontainers.org/incus/docs/main/

brew cli: https://formulae.brew.sh/formula/incus
- brew install incus

brew server:
https://github.com/beringresearch/macpine/blob/main/docs/docs/incus_macpine.md



terraform: https://github.com/lxc/terraform-provider-incus

CI : https://github.com/cloudbase/garm

https://linuxcontainers.org/incus/docs/main/third_party/


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


