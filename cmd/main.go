package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"res/pkg/vm"
)

func main() {
	var Usage = func() {
		log.Printf("Usage: %s <job.json>", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Usage = Usage
	flag.Parse()
	if flag.NArg() != 1 {
		Usage()
		os.Exit(0)
	}
	fileName := flag.Arg(0)
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()
	jobDecoder := json.NewDecoder(file)
	var query vm.JobQuery
	err = jobDecoder.Decode(&query)
	if err != nil {
		log.Fatalf("failed to decode job: %v", err)
	}
	job, err := vm.NewJob(query)
	err = job.Spawn()
	if err != nil {
		log.Fatalf("failed to spawn job: %v", err)
	}
	code := 0
	err = job.ExecScript()
	if err != nil {
		log.Printf("failed to execute script: %v", err)
		code = 1
	}
	err = job.Kill()
	if err != nil {
		log.Fatalf("failed to kill job: %v", err)
		code = 1
	}
	os.Exit(code)
}
