package bootinfo

import "time"

// this package is used to get the boot info like status of database, redis, nats, etc. were
// they stated with errors or not.

var globalBootInfo BootInfo

type ServiceStatus struct {
	Status    string                 `json:"status,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type BootInfo map[string]*ServiceStatus

func (b BootInfo) SetServiceStatus(serviceName string, serviceStatus ServiceStatus) {
	b[serviceName] = &serviceStatus
	serviceStatus.Timestamp = time.Now()
}

func GetBootInfo() BootInfo {
	if globalBootInfo == nil {
		globalBootInfo = make(BootInfo)
	}
	return globalBootInfo
}
