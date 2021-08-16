package main

import (
	"fmt"
	"flag"
	"time"
	"net"
	"context"
	"runtime"
	"runtime/debug"
	"math/rand"
)

var errCount int
var successCount int
//var totalCount int

type config struct {
	resolverCount *int
	duration *int
	memory *bool
	cpu *bool
}

var conf config
const limit = 1127 * 1024 * 1024
func init(){

	const (
		defaultResolverCount = 2
		defaultDuration = 60
	)

	conf.resolverCount =  flag.Int("r", defaultResolverCount, "The number of resolver routines to run, default: 2")
	conf.duration =  flag.Int("d", defaultDuration, "The duration of the test, default: 60")
	conf.memory =  flag.Bool("m", false, "run memory tests, default: false")
	conf.cpu =  flag.Bool("c", false, "run cpu intensive, default: false")
	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.Parse()


	var meg []int32
	if *conf.memory {
		aboutAGig := (1024 * 1024 * 1024) /4
		meg = make([]int32, aboutAGig)

		for i := range meg {
			meg[i] = int32(1)
		}
	}

	printMemUsage()

	tock := time.NewTicker(time.Duration(10) * time.Second)
	defer tock.Stop()
	done := make(chan bool)

	if *conf.cpu {
		rand.Seed(time.Now().UnixNano())
		for x := 0; x < runtime.NumCPU(); x++ {
			go cpuStress(done)
			go cpuStress(done)
			go cpuStress(done)
			go cpuStress(done)
		}
	}
	go func(m []int32) {
		for {
			select {
			case <-done:
				break
			case t := <-tock.C:
				pct := errCount/(errCount + successCount)
				if *conf.memory {
					m[rand.Int31n(35000)] = rand.Int31n(35000)
				}
				fmt.Printf("%s err: %d\t suc: %d\t errPct: %d\n%s ", t, errCount, successCount, pct, t)
				printMemUsage()
//				debug.FreeOSMemory()
			}
		}
	}(meg)
	defer func(){
	done <-true
	}()

	for i := 0; i < *conf.resolverCount; i++ {
		go resolv(done)
	}

	printMemUsage()

	time.Sleep(time.Duration(*conf.duration) * time.Second)
	pct := errCount/(errCount + successCount)
	fmt.Printf("err: %d\t suc: %d\t errPct: %d\n", errCount, successCount, pct)
}

func resolv(d chan bool) {
	ctx := context.TODO()
	for {
		select {
		case <-d:
			break
		default:
			s := time.Now()
			if _, e := net.DefaultResolver.LookupHost(ctx, "github.com"); e != nil { errCount++ } else { successCount++ }
			// We should get about 10/sec/resolver
			e := time.Now()
			if e.Sub(s).Seconds() >= 5 {
				fmt.Printf("Slow name query: %ss\t", e.Sub(s).Seconds())
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func printMemUsage() {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        // For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("HeapObjects = %v", m.HeapObjects)
        fmt.Printf("\tAlloc = %v MiB ", bToMb(m.Alloc))
        fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
        fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
        fmt.Printf("\tNumGC = %v\n", m.NumGC)
	if m.Sys > limit {
		rem := int(limit - m.Sys)
		fmt.Printf("%d of %d bytes used. %d bytes remaining. Garbage Collecting to stay below limit\n", m.Sys, limit, rem)
		debug.FreeOSMemory()
	}
}

func bToMb(b uint64) uint64 {
    return b / 1024 / 1024
}

func cpuStress(d chan bool) {
	for {
		select {
		case <-d:
			break
		default:
			bigA := ^uint64(0) - rand.Uint64()
			bigB := ^uint64(0) - rand.Uint64()
			if bigB == 0 { bigB++ }
			_u := bigA / bigB
			_u++
		}
	}
}
