package main

import (
	"flag"
	"fmt"
	"sync"
)

func main() {
	targetFlakeAttr := flag.String("f", ".#nixosConfigurations", "input string")
	flag.Parse()
	fmt.Println("Parse completed.")

	args := evalJobArgs{
		attr: *targetFlakeAttr,
	}

	numWorkers := 4
	var wg sync.WaitGroup
	resultChan := make(chan evalResult, 128)
	for i := 0; i < numWorkers; i++ {
		go evalResultHandler(resultChan, &wg)
	}
	fmt.Println("Handler started")
	startEvalJobs(args, resultChan)
	fmt.Println("Starting evaluation.")
	wg.Wait()
}
