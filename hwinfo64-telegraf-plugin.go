package main

import (
	"log"

	"hwinfo64-telegraf-plugin/hwinfo"
)

func main() {
    log.Println("Starting hwinfo64-telegraf-plugin.go")

    shmem, err := hwinfo.ReadSharedMem()
    if err != nil {
        log.Printf("ReadSharedMem failed: %v\n", err)
    }
    
    log.Println("ReadSharedMem success")
    log.Println(shmem.NumSensorElements())
}
