package main
//沒平行的方法
import (
	"fmt"
	"math"
	"time"
	"runtime"

)

const sigma = 0.2

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

func americanOptionPrice(S, K, r, T float64, N int, q float64) float64 {
	dt := T / float64(N)
	u := math.Exp((r - q) * dt + sigma * math.Sqrt(dt))
	d := math.Exp((r - q) * dt - sigma * math.Sqrt(dt))
	p := (math.Exp((r - q) * dt) - d) / (u - d)

	// 初始化只有最後一層的選擇權價格
	optionTree := make([][]float64, N + 1)
	for i := range optionTree {
		optionTree[i] = make([]float64, i + 1)
	}
	

	for j := 0; j <= N; j++ {
		optionTree[N][j] = math.Max(0, S * math.Pow(u, float64(j)) * math.Pow(d, float64(N - j)) - K)
	}


	for i := N - 1; i >= 0; i-- {
		for j := 0; j <= i; j++ {
				earlyExercise := math.Max(0, (S * math.Pow(u, float64(j)) * math.Pow(d, float64(i - j))) - K)
				keep := math.Exp(-r * (T / float64(N))) * (p * optionTree[i + 1][j + 1] + (1 - p) * optionTree[i + 1][j])
				optionTree[i][j] = math.Max(keep, earlyExercise)
			}
		}
		
		printMemUsage()
	return optionTree[0][0]
}

func main() {
	S := 80.0     // 初始資產價格
	K := 100.0    // 履約價
	r := 0.08     // 無風險利率
	T := 3.0      // 到期時間（年）
	N := 11000    // 樹層數
	q := 0.12     // 股利率

	startTime := time.Now()
	optionPrice := americanOptionPrice(S, K, r, T, N, q)
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Printf("Total execution time: %v seconds\n", duration.Seconds())
	fmt.Printf("American Option Price: %f\n", optionPrice)
}
