package main

import (
	"encoding/json"
	"fmt"
	"os"

	"AIComputingNode/pkg/hardware"

	"github.com/jaypipes/ghw"
	psutil "github.com/shirou/gopsutil/v3/host"
)

func main() {
	cpu, err := ghw.CPU()
	if err != nil {
		fmt.Printf("Error getting CPU info: %v\n", err)
	} else {
		fmt.Printf("%v\n", cpu)

		for _, processor := range cpu.Processors {
			fmt.Printf("CPU %v: %v (%d, %d)\n", processor.ID, processor.Model,
				processor.NumCores, processor.NumThreads)
		}
	}

	memory, err := ghw.Memory()
	if err != nil {
		fmt.Printf("Error getting memory info: %v\n", err)
	} else {
		fmt.Println(memory.String())
	}

	block, err := ghw.Block()
	if err != nil {
		fmt.Printf("Error getting block storage info: %v\n", err)
	} else {
		fmt.Printf("%v\n", block)

		for _, disk := range block.Disks {
			fmt.Printf(" %v\n", disk)
			for _, part := range disk.Partitions {
				fmt.Printf("  %v\n", part)
			}
		}
	}

	gpu, err := ghw.GPU()
	if err != nil {
		fmt.Printf("Error getting GPU info: %v\n", err)
	} else {
		fmt.Printf("%v\n", gpu)

		for _, card := range gpu.GraphicsCards {
			fmt.Printf(" %v\n", card)
		}
	}

	hi, err := hardware.GetHostInfo()
	if err != nil {
		fmt.Printf("Error getting host info: %v\n", err)
	}
	jsonData, err := json.MarshalIndent(hi, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling host json: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))

	hostInfo, err := psutil.Info()
	if err != nil {
		fmt.Printf("Error retrieving OS info: %v\n", err)
		return
	} else {
		fmt.Printf("OS: %s\n", hostInfo.OS)
		fmt.Printf("Platform: %s\n", hostInfo.Platform)
		fmt.Printf("Platform Family: %s\n", hostInfo.PlatformFamily)
		fmt.Printf("Platform Version: %s\n", hostInfo.PlatformVersion)
		fmt.Printf("Kernel Version: %s\n", hostInfo.KernelVersion)
		fmt.Printf("Kernel Arch: %s\n", hostInfo.KernelArch)
		fmt.Printf("Virtualization System: %s\n", hostInfo.VirtualizationSystem)
		fmt.Printf("Virtualization Role: %s\n", hostInfo.VirtualizationRole)
	}
}
