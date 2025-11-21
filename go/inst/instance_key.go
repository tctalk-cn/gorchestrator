package inst

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// InstanceKey is an instance indicator, identified by hostname and port
type InstanceKey struct {
	Hostname string
	Port     int
}

var (
	ipv4Regexp         = regexp.MustCompile("^([0-9]+)[.]([0-9]+)[.]([0-9]+)[.]([0-9]+)$")
	ipv4HostPortRegexp = regexp.MustCompile("^([^:]+):([0-9]+)$")
	ipv4HostRegexp     = regexp.MustCompile("^([^:]+)$")
	ipv6HostPortRegexp = regexp.MustCompile("^\\[([:0-9a-fA-F]+)\\]:([0-9]+)$") // e.g. [2001:db8:1f70::999:de8:7648:6e8]:3308
	ipv6HostRegexp     = regexp.MustCompile("^([:0-9a-fA-F]+)$")                // e.g. 2001:db8:1f70::999:de8:7648:6e8
)

const detachHint = "//"

func newInstanceKey(hostname string, port int, resolve bool) (instanceKey *InstanceKey, err error) {
	if hostname == "" {
		return instanceKey, fmt.Errorf("NewResolveInstanceKey: Empty hostname")
	}

	instanceKey = &InstanceKey{Hostname: hostname, Port: port}
	if resolve {
		instanceKey, err = instanceKey.ResolveHostname()
	}
	return instanceKey, err
}

// NewResolveInstanceKeyStrings creates and resolves a new instance key based on string params
func NewResolveInstanceKeyStrings(hostname string, port string) (*InstanceKey, error) {
	return newInstanceKeyStrings(hostname, port, true)
}

// newInstanceKeyStrings
func newInstanceKeyStrings(hostname string, port string, resolve bool) (*InstanceKey, error) {
	if portInt, err := strconv.Atoi(port); err != nil {
		return nil, fmt.Errorf("Invalid port: %s", port)
	} else {
		return newInstanceKey(hostname, portInt, resolve)
	}
}

func (this *InstanceKey) ResolveHostname() (*InstanceKey, error) {
	if !this.IsValid() {
		return this, nil
	}

	hostname, err := ResolveHostname(this.Hostname)
	if err == nil {
		this.Hostname = hostname
	}
	return this, err
}

// IsDetached returns 'true' when this hostname is logically "detached"
func (this *InstanceKey) IsDetached() bool {
	return strings.HasPrefix(this.Hostname, detachHint)
}

// IsValid uses simple heuristics to see whether this key represents an actual instance
func (this *InstanceKey) IsValid() bool {
	if this.Hostname == "_" {
		return false
	}
	if this.IsDetached() {
		return false
	}
	return len(this.Hostname) > 0 && this.Port > 0
}
