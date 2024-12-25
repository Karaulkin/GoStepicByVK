package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	md5Semaphore = make(chan struct{}, 1)
)

func semaphoreMd5(val string) string {
	md5Semaphore <- struct{}{}
	defer func() {
		<-md5Semaphore
	}()

	return DataSignerMd5(val)
}

func ExecutePipeline(pipeline ...job) {
	in := make(chan interface{})
	wg := &sync.WaitGroup{}

	for _, task := range pipeline {
		out := make(chan interface{})
		wg.Add(1)
		go func(task job, in, out chan interface{}) {
			defer wg.Done()
			task(in, out)
			close(out)
		}(task, in, out)
		in = out
	}

	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var data []string

	for val := range in {
		if n, ok := val.(string); ok {
			data = append(data, n)
		}
	}

	sort.Strings(data)

	result := strings.Join(data, "_")

	out <- result
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go func(data interface{}) {
			defer wg.Done()

			strData := strconv.Itoa(data.(int))
			crc32Chan := make(chan string)
			crc32md5Chan := make(chan string)

			go func() {
				crc32Chan <- DataSignerCrc32(strData)
				close(crc32Chan)
			}()

			go func() {
				md5Value := semaphoreMd5(strData)
				crc32md5Chan <- DataSignerCrc32(md5Value)
				close(crc32md5Chan)
			}()

			str1 := <-crc32Chan
			str2 := <-crc32md5Chan

			result := str1 + "~" + str2

			out <- result
		}(data)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go func(data string, wg *sync.WaitGroup) {
			defer wg.Done()

			result := make([]string, 6)
			mu := &sync.Mutex{}
			internalWg := &sync.WaitGroup{}

			for i := 0; i < 6; i++ {
				internalWg.Add(1)

				go func(i int, mu *sync.Mutex, wg *sync.WaitGroup) {
					defer internalWg.Done()
					hash := DataSignerCrc32(strconv.Itoa(i) + data)
					mu.Lock()
					result[i] = hash
					mu.Unlock()
				}(i, mu, internalWg)
			}

			internalWg.Wait()
			lipStr := strings.Join(result, "")
			out <- lipStr
		}(data.(string), wg)
	}
	wg.Wait()
}
