package types

import "AIComputingNode/pkg/protocol"

func HostInfo2ProtocolMessage(hostInfo *HostInfo) *protocol.HostInfoResponse {
	res := &protocol.HostInfoResponse{
		Os: &protocol.HostInfoResponse_OSInfo{
			Os:              hostInfo.Os.OS,
			Platform:        hostInfo.Os.Platform,
			PlatformFamily:  hostInfo.Os.PlatformFamily,
			PlatformVersion: hostInfo.Os.PlatformVersion,
			KernelVersion:   hostInfo.Os.KernelVersion,
			KernelArch:      hostInfo.Os.KernelArch,
		},
		Memory: &protocol.HostInfoResponse_MemoryInfo{
			TotalPhysicalBytes: hostInfo.Memory.TotalPhysicalBytes,
			TotalUsableBytes:   hostInfo.Memory.TotalUsableBytes,
		},
	}
	for _, cpu := range hostInfo.Cpu {
		res.Cpu = append(res.Cpu, &protocol.HostInfoResponse_CpuInfo{
			ModelName:    cpu.ModelName,
			TotalCores:   cpu.Cores,
			TotalThreads: cpu.Threads,
		})
	}
	for _, disk := range hostInfo.Disk {
		res.Disk = append(res.Disk, &protocol.HostInfoResponse_DiskInfo{
			DriveType:    disk.DriveType,
			SizeBytes:    disk.SizeBytes,
			Model:        disk.Model,
			SerialNumber: disk.SerialNumber,
		})
	}
	for _, gpu := range hostInfo.Gpu {
		res.Gpu = append(res.Gpu, &protocol.HostInfoResponse_GpuInfo{
			Vendor:  gpu.Vendor,
			Product: gpu.Product,
		})
	}
	return res
}

func ProtocolMessage2HostInfo(res *protocol.HostInfoResponse) *HostInfo {
	hostInfo := &HostInfo{
		Os: OSInfo{
			OS:              res.Os.Os,
			Platform:        res.Os.Platform,
			PlatformFamily:  res.Os.PlatformFamily,
			PlatformVersion: res.Os.PlatformVersion,
			KernelVersion:   res.Os.KernelVersion,
			KernelArch:      res.Os.KernelArch,
		},
		Memory: MemoryInfo{
			TotalPhysicalBytes: res.Memory.TotalPhysicalBytes,
			TotalUsableBytes:   res.Memory.TotalUsableBytes,
		},
	}
	for _, cpu := range res.Cpu {
		hostInfo.Cpu = append(hostInfo.Cpu, CpuInfo{
			ModelName: cpu.ModelName,
			Cores:     cpu.TotalCores,
			Threads:   cpu.TotalThreads,
		})
	}
	for _, disk := range res.Disk {
		hostInfo.Disk = append(hostInfo.Disk, DiskInfo{
			DriveType:    disk.DriveType,
			SizeBytes:    disk.SizeBytes,
			Model:        disk.Model,
			SerialNumber: disk.SerialNumber,
		})
	}
	for _, gpu := range res.Gpu {
		hostInfo.Gpu = append(hostInfo.Gpu, GpuInfo{
			Vendor:  gpu.Vendor,
			Product: gpu.Product,
		})
	}
	return hostInfo
}

func AIProject2ProtocolMessage(projs []AIProjectOfNode) *protocol.AIProjectResponse {
	res := &protocol.AIProjectResponse{}
	for _, proj := range projs {
		res.Projects = append(res.Projects, &protocol.AIProjectOfNode{
			Project: proj.Project,
			Models:  proj.Models,
		})
	}
	return res
}

func ProtocolMessage2AIProject(res *protocol.AIProjectResponse) []AIProjectOfNode {
	projects := make([]AIProjectOfNode, len(res.Projects))
	for i, project := range res.Projects {
		projects[i].Project = project.Project
		projects[i].Models = project.Models
	}
	return projects
}
