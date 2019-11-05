package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"text/tabwriter"
	"time"
)

const max = 200
const divider = 1      //default - 10
const modifier = 2     //default - 2
const monitorSize = 10 //default - 10
const threadsCount = 4
const file = 3

//Thing is for storing string, int and float values in single structure
//Result value is added later to store filter parameter
type Thing struct {
	Company string  `json:"company"`
	Count   int     `json:"count"`
	Price   float32 `json:"price"`
	Result  int
}

//Array of things
type Array struct {
	Things []Thing
	Count  int
}

func (array *Array) add(thing Thing) {
	var i int
	for i = int(array.Count - 1); i >= 0 && array.Things[i].Result > thing.Result; i-- {
		(*array).Things[i+1] = (*array).Things[i]
	}
	(*array).Things[i+1] = thing
	(*array).Count++
}

func mainThread() {
	start := time.Now()
	var wg sync.WaitGroup
	var data []Thing
	data = readJSON(fmt.Sprintf("IFF77_GostautasM_L1_dat_%d.json", file))

	workerch := make(chan Thing, 10)
	resultsch := make(chan Thing)
	datach := make(chan Thing)
	mainch := make(chan Array)
	//donech := make(chan int)

	var workersWaitGroup sync.WaitGroup
	for i := 0; i < threadsCount; i++ {
		workersWaitGroup.Add(1)
		go workerThread(&workersWaitGroup, i, workerch, resultsch)
	}
	wg.Add(2)
	go resultsThread(&wg, resultsch, mainch)
	go dataThread(&wg, datach, workerch)

	for _, thing := range data {
		datach <- thing
	}
	close(datach)
	workersWaitGroup.Wait()
	close(resultsch)

	var results Array
	for result := range mainch {
		results = result
	}

	fmt.Println(time.Since(start))
	fmt.Println("all - ", len(data), "; filtered - ", results.Count)

	printToFile(data, results)
}

func workerThread(wg *sync.WaitGroup, id int, ch <-chan Thing, results chan<- Thing) {
	for value := range ch {
		primeCount := filterCondition(value)

		fmt.Println("gija - ", id, ", kompanija - ", value.Company, " uzsakymo grupavimo indeksas - ", primeCount)

		if (primeCount % modifier) == 0 {
			value.Result = primeCount
			results <- value
		}
	}
	wg.Done()
}

func dataThread(wg *sync.WaitGroup, data <-chan Thing, worker chan<- Thing) {
	for value := range data {
		//fmt.Println(value.Company)
		worker <- value
	}
	close(worker)
	wg.Done()
}

func resultsThread(wg *sync.WaitGroup, resultsch <-chan Thing, mainch chan<- Array) {
	results := Array{Things: make([]Thing, max)}
	for value := range resultsch {
		results.add(value)
	}
	for i := 0; i < results.Count; i++ {
		fmt.Println(results.Things[i].Company)
	}

	mainch <- results
	close(mainch)
	wg.Done()
}

func readJSON(path string) []Thing {
	file, _ := ioutil.ReadFile(path)
	var data []Thing
	_ = json.Unmarshal([]byte(file), &data)
	return data
}

func primeCount(number int) int {
	var n = 3
	var is = true
	var count int

	for n < number {
		is = true
		if n%2 == 0 {
			is = false
		}
		for i := 3; i < (n / 3); i += 2 {
			if (n % i) == 0 {
				is = false
				break
			}
		}
		if is {
			count++
		}
		n++
	}
	return count
}

func filterCondition(thing Thing) int {
	sum := 0
	for _, value := range thing.Company {
		sum += int(value)
	}
	sum += int(thing.Price)
	sum = sum * thing.Count / divider
	primeCount := primeCount(sum)
	return primeCount
}

func printToFile(data []Thing, result Array) {
	f, _ := os.Create("./IFF77_GostautasM_L1_rez.txt")
	w := tabwriter.NewWriter(f, 0, 0, 4, ' ', tabwriter.AlignRight|tabwriter.Debug)

	defer f.Close()

	fmt.Fprintln(w, "rezultatai")
	fmt.Fprintln(w, "-----------\t----------\t---------------\t-------------\t")
	fmt.Fprintln(w, "pavadinimas\tkiekis\tkaina\trezultatas\t")
	fmt.Fprintln(w, "-----------\t----------\t---------------\t-------------\t")

	for i := 0; i < int(result.Count); i++ {
		s := fmt.Sprintf("%s\t%d\t%f\t%d\t", result.Things[i].Company, result.Things[i].Count, result.Things[i].Price, result.Things[i].Result)
		fmt.Fprintln(w, s)
	}

	fmt.Fprintln(w, "duomenys")
	fmt.Fprintln(w, "-----------\t----------\t---------------\t")
	fmt.Fprintln(w, "pavadinimas\tkiekis\tkaina\t")
	fmt.Fprintln(w, "-----------\t----------\t---------------\t")
	for _, value := range data {
		s := fmt.Sprintf("%s\t%d\t%f\t", value.Company, value.Count, value.Price)
		fmt.Fprintln(w, s)
	}

	w.Flush()
}

func main() {
	mainThread()
}
