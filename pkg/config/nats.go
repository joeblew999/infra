package config

import (
	"net"
	"net/url"
	"os"
	"path/filepath"
)

const (
	EnvVarNATSHost = "NATS_HOST"

	NATSLogStreamName    = "LOGS"
	NATSLogStreamSubject = "logs.app"

	NATSClusterNameLocal      = "infra-local"
	NATSClusterNameProduction = "infra-cluster"
	NATSDockerImage           = "nats:alpine"

	NATSClusterBasePort   = 4222
	NATSClusterBaseCPort  = 6222
	NATSClusterBaseHTTP   = 8222
	NATSClusterNodeCount  = 6
	NATSLeafPortOffset    = 100
	NATSLeafPortOffsetDev = 200

	NATSOperatorName        = "infra"
	NATSSystemAccountName   = "SYS"
	NATSApplicationAccount  = "infra"
	NATSApplicationUserName = "infra"
	NATSSystemUserName      = "sys"

	defaultNATSHost = "localhost"
)

func GetNATSPort() string {
	return defaultPort(PortNATS)
}

func GetNatsS3Port() string {
	return defaultPort(PortNatsS3)
}

func GetNATSHost() string {
	if host := os.Getenv(EnvVarNATSHost); host != "" {
		return host
	}
	return defaultNATSHost
}

func GetNATSURL() string {
	return (&url.URL{Scheme: "nats", Host: net.JoinHostPort(GetNATSHost(), GetNATSPort())}).String()
}

func GetNATSS3Endpoint() string {
	return (&url.URL{Scheme: "http", Host: net.JoinHostPort(GetNATSHost(), GetNatsS3Port())}).String()
}

func GetNATSClusterDataPath() string {
	if IsTestEnvironment() {
		return filepath.Join(GetTestDataPath(), "nats-cluster")
	}
	return filepath.Join(GetDataPath(), "nats-cluster")
}

func getNATSAuthBasePath() string {
	if IsTestEnvironment() {
		return filepath.Join(GetTestDataPath(), "nats-auth")
	}
	return filepath.Join(GetDataPath(), "nats-auth")
}

// GetNATSAuthStorePath returns the path used for NSC store material.
func GetNATSAuthStorePath() string {
	return filepath.Join(getNATSAuthBasePath(), "store")
}

// GetNATSAuthCredsPath returns the directory for generated credential files.
func GetNATSAuthCredsPath() string {
	return filepath.Join(getNATSAuthBasePath(), "creds")
}

// GetNATSSystemCredsPath returns the creds file for the system user.
func GetNATSSystemCredsPath() string {
	return filepath.Join(GetNATSAuthCredsPath(), NATSSystemUserName+".creds")
}

// GetNATSApplicationCredsPath returns the creds file for the application user.
func GetNATSApplicationCredsPath() string {
	return filepath.Join(GetNATSAuthCredsPath(), NATSApplicationUserName+".creds")
}

// GetNATSOperatorJWTPath returns the operator JWT path.
func GetNATSOperatorJWTPath() string {
	return filepath.Join(GetNATSAuthStorePath(), NATSOperatorName, NATSOperatorName+".jwt")
}

// GetNATSAccountJWTPath returns the account JWT path for the provided account name.
func GetNATSAccountJWTPath(account string) string {
	return filepath.Join(GetNATSAuthStorePath(), NATSOperatorName, "accounts", account, account+".jwt")
}

func GetNATSClusterName() string {
	if IsProduction() {
		return NATSClusterNameProduction
	}
	return NATSClusterNameLocal
}

func GetNATSDockerImage() string {
	return NATSDockerImage
}

func GetFlyRegions() []string {
	return []string{"iad", "lhr", "nrt", "syd", "fra", "sjc"}
}

// GetUsePillowHubAndSpoke returns whether to use hub-and-spoke topology (true)
// or full mesh clustering (false) for Pillow NATS clusters on Fly.io
func GetUsePillowHubAndSpoke() bool {
	// Default to false (full mesh/FlyioClustering)
	// Set NATS_PILLOW_HUB_AND_SPOKE=true to enable hub-and-spoke topology
	return os.Getenv("NATS_PILLOW_HUB_AND_SPOKE") == "true"
}

func GetNATSClusterNodeCount() int {
	return NATSClusterNodeCount
}

func GetNATSClusterPortsForNode(nodeIndex int) (client, cluster, http, leaf int) {
	client = NATSClusterBasePort + nodeIndex
	http = NATSClusterBaseHTTP + nodeIndex
	leafOffset := NATSLeafPortOffset
	if !IsProduction() {
		leafOffset += NATSLeafPortOffsetDev
	}
	leaf = NATSClusterBasePort + leafOffset + nodeIndex
	if IsProduction() {
		cluster = NATSClusterBaseCPort
	} else {
		cluster = NATSClusterBaseCPort + nodeIndex
	}
	return
}
