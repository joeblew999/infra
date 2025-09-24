package fly

import "fmt"

// GetNATSClusterTemplate returns the fly.toml template for NATS cluster nodes
func GetNATSClusterTemplate(appName, region string) string {
	return fmt.Sprintf(`
app = "%s"
primary_region = "%s"

[build]
  builder = "paketobuildpacks/builder:base"
  buildpacks = ["gcr.io/paketo-buildpacks/go"]

[env]
  ENVIRONMENT = "production"

[processes]
  app = "./main"

[[services]]
  processes = ["app"]
  internal_port = 4222
  protocol = "tcp"
  auto_stop_machines = false
  auto_start_machines = true

  [[services.ports]]
    port = 4222

[[services]]
  processes = ["app"]
  internal_port = 8222
  protocol = "tcp"
  auto_stop_machines = false
  auto_start_machines = true

  [[services.ports]]
    port = 8222
    handlers = ["http"]

[mounts]
  source = "nats_data"
  destination = "/data"

[[vm]]
  memory = 512
  cpu_kind = "shared"
  cpus = 1
`, appName, region)
}

// GetAppServerTemplate returns the fly.toml template for application servers
func GetAppServerTemplate(appName, region string) string {
	return fmt.Sprintf(`
app = "%s"
primary_region = "%s"

[build]
  builder = "paketobuildpacks/builder:base"
  buildpacks = ["gcr.io/paketo-buildpacks/go"]

[env]
  ENVIRONMENT = "production"

[processes]
app = "./main server"

[[services]]
  internal_port = 8080
  protocol = "tcp"
  auto_stop_machines = false
  auto_start_machines = true

  [[services.ports]]
    port = 80
    handlers = ["http"]

  [[services.ports]]
    port = 443
    handlers = ["http", "tls"]

[[vm]]
  memory = 512
  cpu_kind = "shared"
  cpus = 1
`, appName, region)
}

// GetBasicTemplate returns a basic fly.toml template for general use
func GetBasicTemplate(appName, region string) string {
	return fmt.Sprintf(`
app = "%s"
primary_region = "%s"

[build]
  builder = "paketobuildpacks/builder:base"
  buildpacks = ["gcr.io/paketo-buildpacks/go"]

[env]
  ENVIRONMENT = "production"

[[services]]
  internal_port = 8080
  protocol = "tcp"

  [[services.ports]]
    port = 80
    handlers = ["http"]

  [[services.ports]]
    port = 443
    handlers = ["http", "tls"]

[[vm]]
  memory = 256
  cpu_kind = "shared"
  cpus = 1
`, appName, region)
}