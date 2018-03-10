package main

import (
	"flag"
	"fmt"
	"github.com/cunnie/gobonniego/bench"
	"github.com/cunnie/gobonniego/mem"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
)

const Version = "1.0.4"

func main() {
	var verbose, version bool
	var err error
	var numberOfRuns int

	bm := bench.Mark{}

	bonnieParentDir, err := ioutil.TempDir("", "gobonniegoParent")
	check(err)
	defer os.RemoveAll(bonnieParentDir)

	bm.PhysicalMemory, err = mem.Get()
	check(err)

	flag.BoolVar(&verbose, "v", false,
		"Verbose. Will print to stderr diagnostic information such as the amount of RAM, number of cores, etc.")
	flag.BoolVar(&version, "version", false,
		"Version. Will print the current version of gobonniego and then exit")
	flag.IntVar(&numberOfRuns, "runs", 1,
		"The number of test runs")
	flag.IntVar(&bm.NumReadersWriters, "threads", runtime.NumCPU(),
		"The number of concurrent readers/writers, defaults to the number of CPU cores")
	flag.Float64Var(&bm.AggregateTestFilesSizeInGiB, "size", math.Floor(float64(2*int(bm.PhysicalMemory>>20)))/1024,
		"The amount of disk space to use (in GiB), defaults to twice the physical RAM")
	flag.StringVar(&bonnieParentDir, "dir", bonnieParentDir,
		"The directory in which gobonniego places its temporary files, should have at least '-size' space available")
	flag.Parse()

	if version {
		fmt.Printf("gobonniego version %s\n", Version)
		os.Exit(0)
	}

	check(bm.SetBonnieDir(bonnieParentDir))
	defer os.RemoveAll(bm.BonnieDir)

	log.Printf("gobonniego starting. version: %s, threads: %d, disk space to use (MiB): %d",
		Version, bm.NumReadersWriters, int(bm.AggregateTestFilesSizeInGiB*(1<<10)))
	if verbose {
		log.Printf("Number of CPU cores: %d", runtime.NumCPU())
		log.Printf("Total system RAM (MiB): %d", bm.PhysicalMemory>>20)
		log.Printf("Bonnie working directory: %s", bonnieParentDir)
	}

	check(bm.CreateRandomBlock())

	for i := 0; i < numberOfRuns; i++ {
		check(bm.RunSequentialWriteTest())
		if verbose {
			log.Printf("Written (MiB): %d\n", bm.Result[i].WrittenBytes>>20)
			log.Printf("Written (MB): %f\n", float64(bm.Result[i].WrittenBytes)/1000000)
			log.Printf("Duration (seconds): %f\n", bm.Result[i].WrittenDuration.Seconds())
		}
		fmt.Printf("Sequential Write MB/s: %0.2f\n",
			float64(bm.Result[i].WrittenBytes)/float64(bm.Result[i].WrittenDuration.Seconds())/1000000)

		check(bm.RunSequentialReadTest())
		if verbose {
			log.Printf("Read (MiB): %d\n", bm.Result[i].ReadBytes>>20)
			log.Printf("Read (MB): %f\n", float64(bm.Result[i].ReadBytes)/1000000)
			log.Printf("Duration (seconds): %f\n", bm.Result[i].ReadDuration.Seconds())
		}
		fmt.Printf("Sequential Read MB/s: %0.2f\n",
			float64(bm.Result[i].ReadBytes)/float64(bm.Result[i].ReadDuration.Seconds())/1000000)

		check(bm.RunIOPSTest())
		if verbose {
			log.Printf("operations %d\n", bm.Result[i].IOPSOperations)
			log.Printf("Duration (seconds): %f\n", bm.Result[i].IOPSDuration.Seconds())
		}
		fmt.Printf("IOPS: %0.0f\n",
			float64(bm.Result[i].IOPSOperations)/float64(bm.Result[i].IOPSDuration.Seconds()))
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}