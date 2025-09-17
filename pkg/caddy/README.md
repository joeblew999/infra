# Caddy Helpers

- Generate a `Caddyfile` with presets (`PresetSimple`, `PresetDevelopment`, `PresetFull`, `PresetMicroservices`).
- Files land in `.data/caddy/Caddyfile`; ready for goreman or `caddy run --config ...`.
- Call `StartSupervised()` to keep Caddy under goreman supervision.
