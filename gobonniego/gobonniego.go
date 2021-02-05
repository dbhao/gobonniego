package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dbhao/gobonniego/bench"
	"github.com/dbhao/gobonniego/mem"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
	"time"
)

func main() {
	var jsonOut, verbose, version bool
	var err error
	var numberOfRuns, numberOfSecondsToRun int

	bm := bench.Mark{Start: time.Now()}

	bonnieParentDir, err := ioutil.TempDir("", "gobonniegoParent")
	check(err)
	defer os.RemoveAll(bonnieParentDir)

	bm.PhysicalMemory, err = mem.Get()
	check(err)

	flag.BoolVar(&verbose, "v", false,
		"Verbose. Will print to stderr diagnostic information such as the amount of RAM, number of cores, etc.")
	flag.BoolVar(&version, "version", false,
		"Version. Will print the current version of gobonniego and then exit")
	flag.BoolVar(&jsonOut, "json", false,
		"Version. Will print JSON-formatted results to stdout. Does not affect diagnostics to stderr")
	flag.IntVar(&numberOfRuns, "runs", 1,
		"The number of test runs")
	flag.IntVar(&numberOfSecondsToRun, "seconds", 0,
		"The time (in seconds) to run the test")
	flag.IntVar(&bm.NumReadersWriters, "threads", runtime.NumCPU(),
		"The number of concurrent readers/writers, defaults to the number of CPU cores")
	flag.Float64Var(&bm.AggregateTestFilesSizeInGiB, "size", math.Floor(float64(2*int(bm.PhysicalMemory>>20)))/1024,
		"The amount of disk space to use (in GiB), defaults to twice the physical RAM")
	flag.Float64Var(&bm.IODuration, "iops-duration", 15.0,
		"The duration in seconds to run the IOPS benchmark, set to 0.5 for quick feedback during development")
	flag.StringVar(&bonnieParentDir, "dir", bonnieParentDir,
		"The directory in which gobonniego places its temporary files, should have at least '-size' space available")
	flag.Parse()

	if version {
		fmt.Printf("gobonniego version %s\n", bm.Version())
		os.Exit(0)
	}

	check(bm.SetBonnieDir(bonnieParentDir))
	defer os.RemoveAll(bm.BonnieDir)

	log.Printf("gobonniego starting. version: %s, runs: %d, seconds: %d, threads: %d, disk space to use (MiB): %d",
		bm.Version(), numberOfRuns, numberOfSecondsToRun, bm.NumReadersWriters, int(bm.AggregateTestFilesSizeInGiB*(1<<10)))
	if verbose {
		log.Printf("Number of CPU cores: %d", runtime.NumCPU())
		log.Printf("Total system RAM (MiB): %d", bm.PhysicalMemory>>20)
		log.Printf("Bonnie working directory: %s", bonnieParentDir)
	}

	check(bm.CreateRandomBlock())
	go bench.ClearBufferCacheEveryThreeSeconds() // flush the Buffer Cache every 3 seconds

	finishTime := bm.Start.Add(time.Duration(numberOfSecondsToRun) * time.Second)
	for i := 0; (i < numberOfRuns) || time.Now().Before(finishTime); i++ {
		check(bm.RunSequentialWriteTest())
		if verbose {
			log.Printf("Written (MiB): %d\n", bm.Results[i].WrittenBytes>>20)
			log.Printf("Written (MB): %f\n", float64(bm.Results[i].WrittenBytes)/1000000)
			log.Printf("Duration (seconds): %f\n", bm.Results[i].WrittenDuration.Seconds())
		}
		if !jsonOut {
			fmt.Printf("Sequential Write MB/s: %0.2f\n",
				bench.MegaBytesPerSecond(bm.Results[i].WrittenBytes, bm.Results[i].WrittenDuration))
		}

		check(bm.RunSequentialReadTest())
		if verbose {
			log.Printf("Read (MiB): %d\n", bm.Results[i].ReadBytes>>20)
			log.Printf("Read (MB): %f\n", float64(bm.Results[i].ReadBytes)/1000000)
			log.Printf("Duration (seconds): %f\n", bm.Results[i].ReadDuration.Seconds())
		}
		if !jsonOut {
			fmt.Printf("Sequential Read MB/s: %0.2f\n",
				bench.MegaBytesPerSecond(bm.Results[i].ReadBytes, bm.Results[i].ReadDuration))
		}

		check(bm.RunIOPSTest())
		if verbose {
			log.Printf("operations %d\n", bm.Results[i].IOOperations)
			log.Printf("Duration (seconds): %f\n", bm.Results[i].IODuration.Seconds())
		}
		if !jsonOut {
			fmt.Printf("IOPS: %0.0f\n",
				bench.IOPS(bm.Results[i].IOOperations, bm.Results[i].IODuration))
		}
	}
	if jsonOut {
		json.NewEncoder(os.Stdout).Encode(bm)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
