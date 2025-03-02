package main
//改了一下不要一開始就初始化最後一排
//二維沒重複
import (
	"fmt"
	"math"
	"sync"
	"time"
	"runtime"
)

var optionPrice [][]float64

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// 輸出目前分配的記憶體，單位是MB
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// x傳入三角形下面左邊的點
func stencilTriangle(u, d, p, S, K, r, T float64, N int, depth int, y int, x int){

	row := depth
	for i := y; i > y-depth; i-- { //最後一層已經初始化，所以從倒數第二層開始
		for j := x; j < x+row; j++ {
			if i == N-1{
				earlyExercise := math.Max(0, S*math.Pow(u, float64(j))*math.Pow(d, float64(i-j))-K)
				optionPrice[i][j] = earlyExercise
			}else{
				earlyExercise := math.Max(0, S*math.Pow(u, float64(j))*math.Pow(d, float64(i-j))-K)
				keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[i+1][j+1] + (1-p)*optionPrice[i+1][j])
				optionPrice[i][j] = math.Max(keep, earlyExercise)
			}
		}
		row--
	}
}



func stencilRhombus(u, d, p, S, K, r, T float64, N int, depth int, y int, x int) {

	//先做下半三角
	count := 0                     //從最下面只有一個點的開始
	for i := y; i > y-depth; i-- { //從下往上的曾數
		//for j := x; j > x-count; j-- { //每層的點數從1個點的往上跑 從x的點往左
		for j := x-count; j <= x; j++ { 
			earlyExercise := math.Max(0, S*math.Pow(u, float64(j))*math.Pow(d, float64(i-j))-K)
			keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[i+1][j+1] + (1-p)*optionPrice[i+1][j])
			
			optionPrice[i][j] = math.Max(keep, earlyExercise)
			

		}
		count++
	}

	//再做上半三角
	stencilTriangle(u, d, p, S, K, r, T, N, depth-1, y-depth, x-depth+1)
}


func americanOptionPrice(S, K, r, q, sigma, T float64, N int, depth int) {
	optionPrice = make([][]float64, N)
	for i := range optionPrice {
		optionPrice[i] = make([]float64, i+1)
	}
	printMemUsage()
	dt := T / float64(N)
	u := math.Exp((r-q)*dt + sigma*math.Sqrt(dt))
	d := math.Exp((r-q)*dt - sigma*math.Sqrt(dt))
	p := (math.Exp((r-q)*dt) - d) / (u - d)


	var wg sync.WaitGroup
	for i := 0; i < N; i += depth { // 初始化最下層的一排三角形
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			stencilTriangle(u, d, p, S, K, r, T, N, depth, N-1, i)
		}(i)
	}
	wg.Wait()

	for y := N - 2; y >= (depth-1)*2; y -= depth { //從倒數第二層開始

		for x := depth - 1; x <= y-depth+1; x += depth {

			wg.Add(1)
			go func(y, x int) {
				defer wg.Done()
				stencilRhombus(u, d, p, S, K, r, T, N, depth, y, x)
			}(y, x)
		}
		wg.Wait()
	}
	printMemUsage()






}

func main() {
	S := 80.0  // 初始資產價格
	K := 100.0 // 履約價
	r := 0.08  // 無風險利率
	T := 3.0   // 到期時間（年）
	N := 20000 // 樹層數
	q := 0.12  // 股利率
	sigma := 0.2

	length := 10



	startTime := time.Now()
	americanOptionPrice(S, K, r, q, sigma, T, N, length)
	fmt.Printf("optionPrice: %.6f\n", optionPrice[0][0])
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Printf("Total execution time: %v seconds\n", duration.Seconds())
}
