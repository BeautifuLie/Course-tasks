package main

import (
	"fmt"
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

	for i := range in {
		wg.Add(1)
		data := strconv.Itoa(i.(int))
		md5Data := DataSignerMd5(data)
		go func(data string) {
			defer wg.Done()

			crc32chan := make(chan string)

			go crcParallel(crc32chan, data)

			crc32Md5Data := DataSignerCrc32(md5Data)
			crc32Data := <-crc32chan // важно расположить после расчёта crc32Md5Data,чтобы его не блокировать

			out <- crc32Data + "~" + crc32Md5Data
		}(data)
	}
	wg.Wait()
}
func crcParallel(out chan string, data string) {
	out <- DataSignerCrc32(data)
}

func MultiHash(in, out chan interface{}) {
	const TH = 6
	wg := &sync.WaitGroup{}

	for i := range in {

		wg.Add(1)
		go func(i interface{}) {
			defer wg.Done()
			wg1 := &sync.WaitGroup{}

			res := make([]string, TH) //наполнять слайс ТОЛЬКО через индекс
			for j := 0; j < TH; j++ {
				wg1.Add(1)
				data := strconv.Itoa(j) + i.(string)

				go func(j int, res []string) {
					defer wg1.Done()
					data1 := DataSignerCrc32(data)
					res[j] = data1
				}(j, res)

			}
			wg1.Wait()

			out <- strings.Join(res, "")
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
func main() {
	fmt.Println("aaaa")
}
