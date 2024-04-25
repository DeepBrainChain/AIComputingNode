package host

import (
	"AIComputingNode/pkg/log"

	"github.com/jaypipes/ghw"
	ghw_block "github.com/jaypipes/ghw/pkg/block"
	psutil "github.com/shirou/gopsutil/v3/host"
)

type HostInfo struct {
	Os     OSInfo     `json:"os"`
	Cpu    []CpuInfo  `json:"cpu"`
	Memory MemoryInfo `json:"memory"`
	Disk   []DiskInfo `json:"disk"`
	Gpu    []GpuInfo  `json:"gpu"`
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

func GetHostInfo() (*HostInfo, error) {
	var reterr error = nil

	hostInfo, err := psutil.Info()
	if err != nil {
		log.Logger.Warnf("Error retrieving OS info: %v", err)
		reterr = err
	}

	hi := &HostInfo{
		Os: OSInfo{
			OS:              hostInfo.OS,
			Platform:        hostInfo.Platform,
			PlatformFamily:  hostInfo.PlatformFamily,
			PlatformVersion: hostInfo.PlatformVersion,
			KernelVersion:   hostInfo.KernelVersion,
			KernelArch:      hostInfo.KernelArch,
		},
	}

	cpu, err := ghw.CPU()
	if err != nil {
		log.Logger.Warnf("Error getting CPU info: %v", err)
		reterr = err
	}
	for _, processor := range cpu.Processors {
		hi.Cpu = append(hi.Cpu, CpuInfo{
			ModelName: processor.Model,
			Cores:     processor.NumCores,
			Threads:   processor.NumThreads,
		})
	}

	memory, err := ghw.Memory()
	if err != nil {
		log.Logger.Warnf("Error getting memory info: %v", err)
		reterr = err
	}
	hi.Memory.TotalPhysicalBytes = memory.TotalPhysicalBytes
	hi.Memory.TotalUsableBytes = memory.TotalUsableBytes

	block, err := ghw.Block()
	if err != nil {
		log.Logger.Warnf("Error getting block storage info: %v", err)
		reterr = err
	}
	for _, disk := range block.Disks {
		if disk.StorageController != ghw_block.STORAGE_CONTROLLER_LOOP {
			hi.Disk = append(hi.Disk, DiskInfo{
				DriveType:    disk.DriveType.String(),
				SizeBytes:    disk.SizeBytes,
				Model:        disk.Model,
				SerialNumber: disk.SerialNumber,
			})
		}
	}

	gpu, err := ghw.GPU()
	if err != nil {
		log.Logger.Warnf("Error getting GPU info: %v", err)
		reterr = err
	}
	for _, card := range gpu.GraphicsCards {
		vendor := ""
		product := ""
		if card.DeviceInfo != nil {
			if card.DeviceInfo.Vendor != nil {
				vendor = card.DeviceInfo.Vendor.Name
			}
			if card.DeviceInfo.Product != nil {
				product = card.DeviceInfo.Product.Name
			}
		}
		if product != "" {
			hi.Gpu = append(hi.Gpu, GpuInfo{
				Vendor:  vendor,
				Product: product,
			})
		}
	}

	return hi, reterr
}
