package inst

// InstanceKey is an instance indicator, identified by hostname and port
type InstanceKey struct {
	Hostname string
	Port     int
}
