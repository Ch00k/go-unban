// Package f2bclient provides functions, that implement simple abstractions over `fail2ban-client` command line utility.
package f2bclient

import (
	"os/exec"
	"strings"
)

var execCommand = exec.Command

func runFail2banClient(env []string, args ...string) (string, error) {
	cmd := execCommand("fail2ban-client", args...)
	cmd.Env = append(cmd.Env, env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func getJails() ([]string, error) {
	out, err := runFail2banClient(nil, "status")
	if err != nil {
		return nil, err
	}
	parts := strings.Split(out, "Jail list:")
	jails := strings.TrimSpace(parts[1])
	return strings.Split(jails, ", "), nil
}

func getIPsInJail(jail string) ([]string, error) {
	out, err := runFail2banClient(nil, "status", jail)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(out, "Banned IP list:")
	ips := strings.TrimSpace(parts[1])
	return strings.Split(ips, " "), nil
}

// IsIPBanned checks whether the IP address is present in any of fail2ban existing jails.
// It returns a boolean and any error encountered.
func IsIPBanned(ip string) (bool, error) {
	jails, err := getJails()
	if err != nil {
		return false, err
	}
	for _, jail := range jails {
		ips, err := getIPsInJail(jail)
		if err != nil {
			return false, err
		}
		if contains(ips, ip) {
			return true, nil
		}
	}
	return false, nil
}

// UnbanIP removes the IP address from all fail2ban existing jails.
// Returns the error if encountered.
func UnbanIP(ip string) error {
	_, err := runFail2banClient(nil, "unban", ip)
	return err
}

func contains(array []string, item string) bool {
	for _, i := range array {
		if i == item {
			return true
		}
	}
	return false
}
