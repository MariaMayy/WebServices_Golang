package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код
const goroutinesMaxCount = 100
const cTH = 6 // 6 хешей

var mu sync.Mutex

// ----- ExecutePipeline -----
// ----- SingleHash -----
// ----- MultiHash -----
// ----- CombineResults -----
// crc32 через DataSignerCrc32
// md5 через DataSignerMd5

func SingleHashWorker(in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done() // уменьшаем счетчик

	for input := range in {
		dataHash := fmt.Sprintf("%v", input)
		workerResult := make(chan string)
		str := make(chan string)

		mu.Lock()
		SecondMD5 := DataSignerMd5(dataHash) // т.к может вызываться одновременно только 1 раз
		mu.Unlock()

		go func(dataH string, Result chan string) {
			First := DataSignerCrc32(dataH)
			Result <- First
		}(dataHash, workerResult)

		go func(sMD5 string, Result chan string) {
			Second := DataSignerCrc32(sMD5)
			Result <- Second
		}(SecondMD5, str)

		out <- fmt.Sprintf("%v", <-workerResult) + "~" + fmt.Sprintf("%v", <-str)

	}
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{} // инициализируем группу

	for i := 0; i < goroutinesMaxCount; i++ {
		wg.Add(1) // увеличиваем счетчик
		go SingleHashWorker(in, out, wg)
	}

	wg.Wait() // ждем, когда wg.Done() обнулит счетчик
}

func MultiHashWorkerTH(index int, s string, wgTH *sync.WaitGroup, mapMH map[int]string) {
	defer wgTH.Done()

	sIndex := strconv.Itoa(index)
	MH := DataSignerCrc32(sIndex + s)
	mu.Lock()
	mapMH[index] = MH
	mu.Unlock()
}

// Сортировка map по ключам
func SortKeys(sMap map[int]string) []string {
	keys := make([]int, len(sMap))
	strResult := make([]string, cTH)
	i := 0

	for keyValue := range sMap {
		keys[i] = keyValue
		i++
	}

	sort.Ints(keys)
	for _, k := range keys {
		strResult[k] = sMap[k]
	}

	return strResult
}

func MultiHashWorker(in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done() // уменьшаем счетчик
	mu.Lock()
	mapMH := make(map[int]string, cTH)
	mu.Unlock()
	for input := range in {
		dataHash := fmt.Sprintf("%v", input)
		wgTH := &sync.WaitGroup{}

		for j := 0; j < cTH; j++ {
			wgTH.Add(1)
			go MultiHashWorkerTH(j, dataHash, wgTH, mapMH)
		}

		wgTH.Wait()

		mu.Lock()
		sl := SortKeys(mapMH) // сортируем map
		out <- strings.Join(sl, "")
		mu.Unlock()
	}

}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{} // инициализируем группу

	for i := 0; i < goroutinesMaxCount; i++ {
		wg.Add(1) // увеличиваем счетчик
		go MultiHashWorker(in, out, wg)
	}

	wg.Wait() // ждем, когда wg.Done() обнулит счетчик
}

func CombineResults(in, out chan interface{}) {
	var Result []string

	for input := range in {
		strHash := fmt.Sprintf("%v", input)
		Result = append(Result, strHash)
	}
	sort.Strings(Result)
	sResult := strings.Join(Result, "_")

	out <- sResult
}

func ExecutePipeline(hashSignJobs ...job) {
	wg := &sync.WaitGroup{}
	input := make(chan interface{})

	for _, jobHash := range hashSignJobs {
		wg.Add(1)
		output := make(chan interface{})

		go func(jobH job, in, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()  // уменьшаем счетчик
			defer close(out) // закрываем out
			jobH(in, out)
		}(jobHash, input, output, wg)

		input = output
	}

	wg.Wait()
}

func main() {

}
