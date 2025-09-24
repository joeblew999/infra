package config

import (
	"net"
	"net/url"
)

const defaultLocalHost = "localhost"

// GetLocalHost returns the default hostname for local URLs.
func GetLocalHost() string {
	return defaultLocalHost
}

// FormatLocalHostPort returns host:port for the provided port using the configured local host.
func FormatLocalHostPort(port string) string {
	return net.JoinHostPort(GetLocalHost(), port)
}

// FormatLocalURL builds a local URL with the given scheme and port.
func FormatLocalURL(scheme, port string) string {
	return (&url.URL{Scheme: scheme, Host: FormatLocalHostPort(port)}).String()
}

// FormatLocalHTTP builds an HTTP URL for the given port on the local host.
func FormatLocalHTTP(port string) string {
	return FormatLocalURL("http", port)
}

// FormatLocalHTTPS builds an HTTPS URL for the given port on the local host.
func FormatLocalHTTPS(port string) string {
	return FormatLocalURL("https", port)
}
