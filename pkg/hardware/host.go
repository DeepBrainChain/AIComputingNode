package hardware

import (
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/types"

	"github.com/jaypipes/ghw"
	ghw_block "github.com/jaypipes/ghw/pkg/block"
	psutil "github.com/shirou/gopsutil/v3/host"
)

func GetHostInfo() (*types.HostInfo, error) {
	var reterr error = nil

	hi := &types.HostInfo{}
	hostInfo, err := psutil.Info()
	if err != nil {
		log.Logger.Warnf("Error retrieving OS info: %v", err)
		reterr = err
	} else {
		hi.Os.OS = hostInfo.OS
		hi.Os.Platform = hostInfo.Platform
		hi.Os.PlatformFamily = hostInfo.PlatformFamily
		hi.Os.PlatformVersion = hostInfo.PlatformVersion
		hi.Os.KernelVersion = hostInfo.KernelVersion
		hi.Os.KernelArch = hostInfo.KernelArch
	}

	cpu, err := ghw.CPU()
	if err != nil {
		log.Logger.Warnf("Error getting CPU info: %v", err)
		reterr = err
	} else {
		for _, processor := range cpu.Processors {
			hi.Cpu = append(hi.Cpu, types.CpuInfo{
				ModelName: processor.Model,
				Cores:     processor.NumCores,
				Threads:   processor.NumThreads,
			})
		}
	}

	memory, err := ghw.Memory()
	if err != nil {
		log.Logger.Warnf("Error getting memory info: %v", err)
		reterr = err
	} else {
		hi.Memory.TotalPhysicalBytes = memory.TotalPhysicalBytes
		hi.Memory.TotalUsableBytes = memory.TotalUsableBytes
	}

	block, err := ghw.Block()
	if err != nil {
		log.Logger.Warnf("Error getting block storage info: %v", err)
		reterr = err
	} else {
		for _, disk := range block.Disks {
			if disk.StorageController != ghw_block.STORAGE_CONTROLLER_LOOP {
				hi.Disk = append(hi.Disk, types.DiskInfo{
					DriveType:    disk.DriveType.String(),
					SizeBytes:    disk.SizeBytes,
					Model:        disk.Model,
					SerialNumber: disk.SerialNumber,
				})
			}
		}
	}

	gpu, err := ghw.GPU()
	if err != nil {
		log.Logger.Warnf("Error getting GPU info: %v", err)
		reterr = err
	} else {
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
				hi.Gpu = append(hi.Gpu, types.GpuInfo{
					Vendor:  vendor,
					Product: product,
				})
			}
		}
	}

	return hi, reterr
}
