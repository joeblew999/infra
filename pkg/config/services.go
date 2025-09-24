package config

import "path/filepath"

func GetPocketBaseDataPath() string {
	if IsTestEnvironment() {
		return filepath.Join(GetTestDataPath(), "pocketbase")
	}
	return filepath.Join(GetDataPath(), "pocketbase")
}

func GetPocketBasePort() string {
	return defaultPort(PortPocketBase)
}

func GetBentoPath() string {
	return filepath.Join(GetDataPath(), "bento")
}

func GetBentoPort() string {
	return defaultPort(PortBento)
}

func GetCaddyPath() string {
	if IsTestEnvironment() {
		return filepath.Join(GetTestDataPath(), "caddy")
	}
	return filepath.Join(GetDataPath(), "caddy")
}

func GetCaddyPort() string {
	return defaultPort(PortCaddy)
}

func GetWebServerPort() string {
	return defaultPort(PortWebServer)
}

func GetMCPPort() string {
	return defaultPort(PortMCP)
}

func GetDeckAPIPort() string {
	return defaultPort(PortDeckAPI)
}

func GetMetricsPort() string {
	return defaultPort(PortMetrics)
}

func GetXTemplatePath() string {
	return filepath.Join(GetDataPath(), "xtemplate")
}

func GetXTemplatePort() string {
	return "8080"
}

func GetDeckPath() string {
	if IsTestEnvironment() {
		return filepath.Join(GetTestDataPath(), DeckDir)
	}
	return filepath.Join(GetDataPath(), DeckDir)
}

func GetDeckBinPath() string {
	return filepath.Join(GetDeckPath(), "bin")
}

func GetDeckWASMPath() string {
	return filepath.Join(GetDeckPath(), "wasm")
}

func GetDeckCachePath() string {
	return filepath.Join(GetDeckPath(), "cache")
}

func GetMjmlPath() string {
	if IsTestEnvironment() {
		return filepath.Join(GetTestDataPath(), MjmlDir)
	}
	return filepath.Join(GetDataPath(), MjmlDir)
}

func GetHugoPort() string {
	return defaultPort(PortHugo)
}
