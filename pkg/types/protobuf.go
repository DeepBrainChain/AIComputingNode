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
			OS:              res.Os.GetOs(),
			Platform:        res.Os.GetPlatform(),
			PlatformFamily:  res.Os.GetPlatformFamily(),
			PlatformVersion: res.Os.GetPlatformVersion(),
			KernelVersion:   res.Os.GetKernelVersion(),
			KernelArch:      res.Os.GetKernelArch(),
		},
		Memory: MemoryInfo{
			TotalPhysicalBytes: res.Memory.GetTotalPhysicalBytes(),
			TotalUsableBytes:   res.Memory.GetTotalUsableBytes(),
		},
	}
	for _, cpu := range res.Cpu {
		hostInfo.Cpu = append(hostInfo.Cpu, CpuInfo{
			ModelName: cpu.GetModelName(),
			Cores:     cpu.GetTotalCores(),
			Threads:   cpu.GetTotalThreads(),
		})
	}
	for _, disk := range res.Disk {
		hostInfo.Disk = append(hostInfo.Disk, DiskInfo{
			DriveType:    disk.GetDriveType(),
			SizeBytes:    disk.GetSizeBytes(),
			Model:        disk.GetModel(),
			SerialNumber: disk.GetSerialNumber(),
		})
	}
	for _, gpu := range res.Gpu {
		hostInfo.Gpu = append(hostInfo.Gpu, GpuInfo{
			Vendor:  gpu.GetVendor(),
			Product: gpu.GetProduct(),
		})
	}
	return hostInfo
}

func AIProject2ProtocolMessage(projs []AIProjectConfig, nt uint32) *protocol.AIProjectResponse {
	res := &protocol.AIProjectResponse{
		NodeType: nt,
	}
	for _, proj := range projs {
		models := make([]*protocol.AIModelOfProject, 0)
		for _, model := range proj.Models {
			models = append(models, &protocol.AIModelOfProject{
				Model: model.Model,
				Api:   model.API,
				Type:  uint32(model.Type),
				// Idle:  uint32(model.Type),
			})
		}
		res.Projects = append(res.Projects, &protocol.AIProjectOfNode{
			Project: proj.Project,
			Models:  models,
		})
	}
	return res
}

func ProtocolMessage2AIProject(res *protocol.AIProjectResponse) []AIProjectConfig {
	projects := make([]AIProjectConfig, len(res.Projects))
	for i, project := range res.Projects {
		projects[i].Project = project.GetProject()
		projects[i].Models = make([]AIModelConfig, 0)
		for _, model := range project.Models {
			projects[i].Models = append(projects[i].Models, AIModelConfig{
				Model: model.GetModel(),
				API:   model.GetApi(),
				Type:  int(model.GetType()),
			})
		}
	}
	return projects
}
