package hardware

import (
	"github.com/jaypipes/ghw"
)

type Hardware struct {
	Cpu    []CpuInfo  `json:"cpu"`
	Memory MemoryInfo `json:"memory"`
	Disk   []DiskInfo `json:"disk"`
	Gpu    []GpuInfo  `json:"gpu"`
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

func GetHardwareInfo() (*Hardware, error) {
	hd := &Hardware{}

	cpu, err := ghw.CPU()
	if err != nil {
		return hd, err
	}
	for _, processor := range cpu.Processors {
		hd.Cpu = append(hd.Cpu, CpuInfo{
			ModelName: processor.Model,
			Cores:     processor.NumCores,
			Threads:   processor.NumThreads,
		})
	}

	memory, err := ghw.Memory()
	if err != nil {
		return hd, err
	}
	hd.Memory.TotalPhysicalBytes = memory.TotalPhysicalBytes
	hd.Memory.TotalUsableBytes = memory.TotalUsableBytes

	block, err := ghw.Block()
	if err != nil {
		return hd, err
	}
	for _, disk := range block.Disks {
		hd.Disk = append(hd.Disk, DiskInfo{
			DriveType:    disk.DriveType.String(),
			SizeBytes:    disk.SizeBytes,
			Model:        disk.Model,
			SerialNumber: disk.SerialNumber,
		})
	}

	gpu, err := ghw.GPU()
	if err != nil {
		return hd, err
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
			hd.Gpu = append(hd.Gpu, GpuInfo{
				Vendor:  vendor,
				Product: product,
			})
		}
	}

	return hd, nil
}
