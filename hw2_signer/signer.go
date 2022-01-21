package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(freeFlowJobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, job := range freeFlowJobs {
		wg.Add(1)
		out := make(chan interface{})

		go func(in, out chan interface{}, job func(in, out chan interface{})) {
			defer wg.Done()
			defer close(out)
			job(in, out)
		}(in, out, job)
		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for i := range in {
		wg.Add(1)

		go func(i interface{}) {
			defer wg.Done()
			data := strconv.Itoa(i.(int))

			mu.Lock()
			md5Data := DataSignerMd5(data)
			mu.Unlock()

			crc32Data := DataSignerCrc32(data)
			crc32Md5Data := DataSignerCrc32(md5Data)

			out <- crc32Data + "~" + crc32Md5Data
		}(i)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	const TH = 6
	wg := &sync.WaitGroup{}

	for i := range in {
		result := make([]string, TH)
		wg.Add(1)
		go func(i interface{}) {
			defer wg.Done()

			for j := 0; j < TH; j++ {

				data := strconv.Itoa(j) + i.(string)
				data1 := DataSignerCrc32(data)

				result = append(result, data1)

			}
			out <- strings.Join(result, "")
		}(i)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var result []string

	for i := range in {
		result = append(result, i.(string))
	}

	sort.Strings(result)
	out <- strings.Join(result, "_")
}
