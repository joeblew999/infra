# UTM Helpers

Lightweight helpers for spinning up macOS-hosted VMs (Windows/Linux) so the infra stack can be exercised from non-mac environments. Currently experimental and subject to change.

This will allow use to test the system on Windows and Linux Desktops, to ensure everything works. 

We do not rely on docker, except for Cloud deployments and even then its very much just a wrapper, with the same binary running inside a docker that we run on Desktops and Laptops.
