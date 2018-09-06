package main

import (
	_ "fmt"
	"sort"
	"strconv"
	"strings"
)

func Crc32Parallel(data string, res chan string) {
	res <- DataSignerCrc32(data)
	close(res)
}

func SingleHash(in, out chan interface{}) {
	var res []chan string
	i := 0

	for data := range in {

		stringData := strconv.Itoa(data.(int))
		res = append(res, make(chan string, 1))
		go Crc32Parallel(stringData, res[i])

		res = append(res, make(chan string, 1))
		md5 := DataSignerMd5(stringData)
		go Crc32Parallel(md5, res[i+1])

		i += 2

	}

	for j := 0; j < i; j += 2 {
		crc32 := <-res[j]
		crc32md5 := <-res[j+1]

		out <- crc32 + "~" + crc32md5
	}

	// LOOP:
	//         for {
	//             select {
	//                 case crc32 = <- res1:
	//         		case crc32md5 = <- res2:
	//         		default:
	//         			break LOOP
	//     		}
	//         }
}

func MultiHash(in, out chan interface{}) {
	var res []chan string
	i := 0

	for data := range in {

		for th := 0; th <= 5; th++ {
			res = append(res, make(chan string, 1))
			go Crc32Parallel(strconv.Itoa(th)+data.(string), res[i+th])
		}

		i += 6
	}

	for j := 0; j < i; j += 6 {
		out <- <-res[j] + <-res[j+1] + <-res[j+2] + <-res[j+3] + <-res[j+4] + <-res[j+5]
	}
}

func CombineResults(in, out chan interface{}) {
	var result []string

	for data := range in {
		result = append(result, data.(string))
	}

	sort.Strings(result)
	out <- strings.Join(result, "_")
}

func ExecutePipeline(jobs ...job) {

	in := make(chan interface{}, MaxInputDataLen)
	out := make(chan interface{}, MaxInputDataLen)

	for _, j := range jobs {
		j(in, out)

		in = out
		out = make(chan interface{}, MaxInputDataLen)
		close(in)
	}
}
