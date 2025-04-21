package main

import (
	"encoding/json"
	"flag"
	"os"
	"res/pkg/log"
	"res/pkg/vm"
)

func main() {
	var logger log.Logger
	logger = log.HumanLogger{Out: os.Stdout, Err: os.Stderr}
	var Usage = func() {
		logger.Printf(log.LevelStdout, "Usage: %s <job.json>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Usage = Usage
	jsonOut := flag.Bool("json", false, "output in json format")
	flag.Parse()
	if flag.NArg() != 1 {
		Usage()
		os.Exit(3)
	}
	if *jsonOut {
		logger = log.JsonLogger{Out: os.Stdout}
	}
	fileName := flag.Arg(0)
	file, err := os.Open(fileName)
	if err != nil {
		logger.Printf(log.LevelError, "failed to open file: %v", err)
		os.Exit(2)
	}
	defer func() { _ = file.Close() }()
	jobDecoder := json.NewDecoder(file)
	var query vm.JobQuery
	err = jobDecoder.Decode(&query)
	if err != nil {
		logger.Printf(log.LevelError, "failed to decode job: %v", err)
		os.Exit(2)
	}
	job, err := vm.NewJob(query)
	if err != nil {
		logger.Printf(log.LevelError, "failed to create job: %v", err)
		os.Exit(2)
	}
	err = job.Spawn()
	if err != nil {
		logger.Printf(log.LevelError, "failed to spawn job: %v", err)
		os.Exit(2)
	}
	code := 0
	err = job.ExecScript(logger)
	if err != nil {
		logger.Printf(log.LevelError, "failed to execute script: %v", err)
		code = 1
	}
	err = job.Kill()
	if err != nil {
		logger.Printf(log.LevelError, "failed to kill job: %v", err)
		code = 1
	}
	os.Exit(code)
}
