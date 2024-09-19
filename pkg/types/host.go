package types

type HostInfo struct {
	Os     OSInfo     `json:"os,omitempty"`
	Cpu    []CpuInfo  `json:"cpu,omitempty"`
	Memory MemoryInfo `json:"memory,omitempty"`
	Disk   []DiskInfo `json:"disk,omitempty"`
	Gpu    []GpuInfo  `json:"gpu,omitempty"`
}

type OSInfo struct {
	OS              string `json:"os"`               // ex: freebsd, linux
	Platform        string `json:"platform"`         // ex: ubuntu, linuxmint
	PlatformFamily  string `json:"platform_family"`  // ex: debian, rhel
	PlatformVersion string `json:"platform_version"` // version of the complete OS
	KernelVersion   string `json:"kernel_version"`   // version of the OS kernel (if available)
	KernelArch      string `json:"kernel_arch"`      // native cpu architecture queried at runtime, as returned by `uname -m` or empty string in case of error
}

type CpuInfo struct {
	ModelName string `json:"model_name"`
	Cores     uint32 `json:"total_cores"`
	Threads   uint32 `json:"total_threads"`
}

type MemoryInfo struct {
	TotalPhysicalBytes int64 `json:"total_physical_bytes"`
	TotalUsableBytes   int64 `json:"total_usable_bytes"`
}

type DiskInfo struct {
	DriveType    string `json:"drive_type"`
	SizeBytes    uint64 `json:"size_bytes"`
	Model        string `json:"model"`
	SerialNumber string `json:"serial_number"`
}

type GpuInfo struct {
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
}
