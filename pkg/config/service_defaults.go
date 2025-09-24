package config

// PortKey represents a logical port within the infrastructure stack.
type PortKey string

const (
	PortWebServer  PortKey = "web_server"
	PortPocketBase PortKey = "pocketbase"
	PortNATS       PortKey = "nats"
	PortMCP        PortKey = "mcp"
	PortBento      PortKey = "bento"
	PortCaddy      PortKey = "caddy"
	PortDeckAPI    PortKey = "deck_api"
	PortMetrics    PortKey = "metrics"
	PortHugo       PortKey = "hugo"
	PortNatsS3     PortKey = "nats_s3"
)

var defaultServicePorts = map[PortKey]string{
	PortWebServer:  "1337",
	PortPocketBase: "8090",
	PortNATS:       "4222",
	PortMCP:        "8080",
	PortBento:      "4195",
	PortCaddy:      "80",
	PortDeckAPI:    "8888",
	PortMetrics:    "9091",
	PortHugo:       "1313",
	PortNatsS3:     "5222",
}

func defaultPort(key PortKey) string {
	if port, ok := defaultServicePorts[key]; ok {
		return port
	}
	return ""
}
