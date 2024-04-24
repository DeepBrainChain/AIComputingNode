package main

import (
	"AIComputingNode/pkg/hardware"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jaypipes/ghw"
)

func main() {
	cpu, err := ghw.CPU()
	if err != nil {
		fmt.Printf("Error getting CPU info: %v", err)
	}
	fmt.Printf("%v\n", cpu)

	for _, processor := range cpu.Processors {
		fmt.Printf("CPU %v: %v (%d, %d)\n", processor.ID, processor.Model,
			processor.NumCores, processor.NumThreads)
	}

	memory, err := ghw.Memory()
	if err != nil {
		fmt.Printf("Error getting memory info: %v", err)
	}

	fmt.Println(memory.String())

	block, err := ghw.Block()
	if err != nil {
		fmt.Printf("Error getting block storage info: %v", err)
	}

	fmt.Printf("%v\n", block)

	for _, disk := range block.Disks {
		fmt.Printf(" %v\n", disk)
		for _, part := range disk.Partitions {
			fmt.Printf("  %v\n", part)
		}
	}

	gpu, err := ghw.GPU()
	if err != nil {
		fmt.Printf("Error getting GPU info: %v", err)
	}

	fmt.Printf("%v\n", gpu)

	for _, card := range gpu.GraphicsCards {
		fmt.Printf(" %v\n", card)
	}

	hd, err := hardware.GetHardwareInfo()
	if err != nil {
		fmt.Printf("Error getting hardware info: %v", err)
		os.Exit(1)
	}
	jsonData, err := json.MarshalIndent(hd, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling hardware json: %v", err)
		os.Exit(1)
	}
	fmt.Print(string(jsonData))
}
