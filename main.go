package main

import (
	"fmt"
	"hash/fnv"
	"math"
	"sync"

	"github.com/datadog/hyperloglog"
	"github.com/google/uuid"
)

const elements = 20_000_000
const duplicatesEachIteration = 3
const buckets = 16
const workers = 1

func main() {
	l, err := hyperloglog.New(buckets)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			hasher := fnv.New32()

			for i := 0; i < elements; i++ {
				u, err := uuid.NewRandom()
				if err != nil {
					panic(err)
				}
				hasher.Reset()
				bin, err := u.MarshalBinary()
				if err != nil {
					panic(err)
				}
				hasher.Write(bin)
				for j := 0; j < duplicatesEachIteration; j++ {
					l.Add(hasher.Sum32())
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()

	uniqueElements := elements * workers
	totalElements := uniqueElements * duplicatesEachIteration

	fmt.Printf("approximate number of unique elements: %d\n", l.Count())
	fmt.Printf("inserted %d elements, unique elements %d, correctness: %f\n",
		totalElements,
		uniqueElements,
		(1.0 - (math.Abs(float64(l.Count())-float64(uniqueElements)) / float64(uniqueElements))))
}
