package main

import (
	"github.com/shirou/gopsutil/mem"
	"fmt"
	"github.com/shirou/gopsutil/process"
)

func main() {
	v, _ := mem.VirtualMemory()

	// almost every return value is a struct
	fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

	// convert to JSON. String() is also implemented
	fmt.Println(process.PidExists(9012))
}

