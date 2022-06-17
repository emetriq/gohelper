package env

import "os"

// GetHostname returns the hostname of the current machine.
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = os.Getenv("HOSTNAME")
	}
	return hostname
}
