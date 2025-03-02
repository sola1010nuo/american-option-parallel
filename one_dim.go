package main
//一維存法 減少cache miss 但一樣耗記憶體
import (
	"fmt"
	"math"
	"time"
	"sync"
	"runtime"
)

var N = 20000 // 樹層數
var optionPrice = make([]float64, N*(N+1)/2)
var triangle_block_size int
var rhombus_block_size int

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

func count_rhombus(N int, depth int, times int) int { //計算 (N / depth - 2) + (N / depth -1) + ... times次
	var start = N / depth - 1
	var count int = 0
	for i := 0; i < times ; i++ {
		count += start
		start--
	}
	return count
}

func stencilTriangle(u, d, p, S, K, r, T float64, N int, depth int, start_position int) {
	var now_position int = start_position
	var row int = depth - 1

	for i := 0; i < depth; i++ {	
		upper := now_position % triangle_block_size + (start_position / triangle_block_size) * depth //原本是start是now有問題再改回來		
		down := N - 1 - upper
		optionPrice[now_position] = math.Max(0, S * math.Pow(u, float64(upper) ) * math.Pow(d, float64(down)) - K)
		now_position++
	}

	for i := 0; i < depth - 1; i++ { 
		for j := 0; j < row; j++ {
			//var upper = now_position - count_NowToEnd(depth,row) - start_position + (start_position / block_size) * depth
			var upper = (start_position / triangle_block_size)*depth + j

			var down = (N - 1) - (i+1) - upper
			earlyExercise := math.Max(0, S*math.Pow(u, float64(upper))*math.Pow(d, float64(down)) - K)
			keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[now_position - row] + (1-p)*optionPrice[now_position - row - 1])
			optionPrice[now_position] = math.Max(keep, earlyExercise)
			now_position++
		}
		row--
	}

}

func stencilSecondLayerRhombus(u, d, p, S, K, r, T float64, N int, depth int, start_position int, rhombus_num int ) {
	var now_position int = start_position
	//fmt.Printf("start_position: %v num %v\n", now_position, rhombus_num)

	var left = triangle_block_size * rhombus_num + (depth -1)
	var right = triangle_block_size * (rhombus_num + 1)
	var row = 0

	//倒三角形
	for i := 0; i < depth; i++ {
		for j := 0; j <= row; j++ {
			upper := (rhombus_num * depth) + (depth - 1) - (row - j)
			down := (N - 2) - i - upper
			earlyExercise := math.Max(0, S*math.Pow(u, float64(upper))*math.Pow(d, float64(down)) - K)
			if i == 0 && j == 0{
				keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[right] + (1-p)*optionPrice[left])
				optionPrice[now_position] = math.Max(keep, earlyExercise)
				now_position++
				left += (depth - 1 + i)
				right += (depth - i)
			}else if j == 0{ //最左邊的 用num的右邊
				keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[now_position - i] + (1-p)*optionPrice[left])
				optionPrice[now_position] = math.Max(keep, earlyExercise)
				now_position++
				left += (depth - 1 - i)

			}else if j == row{ //最右邊的 用num +1 的左邊
				keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[right] + (1-p)*optionPrice[now_position - i - 1])
				optionPrice[now_position] = math.Max(keep, earlyExercise)
				now_position++
				right += (depth - i)

			}else { //中間的 用菱形自己的左右
				keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[now_position-i] + (1-p)*optionPrice[now_position - i - 1])
				optionPrice[now_position] = math.Max(keep, earlyExercise)
				now_position++
			}

		}
		row++
	}

	//上面的三角形 大小為depth-1
	row = depth - 1
	for i := 0; i < depth - 1; i++ {
		for j := 0; j < row; j++ {
			upper := (rhombus_num * depth) + j
			down := ((N - 1) - depth - 1) - i -upper // N-1 是總共移動的次數 depth-1是上面的三角形
			earlyExercise := math.Max(0, S*math.Pow(u, float64(upper))*math.Pow(d, float64(down)) - K)
			keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[now_position - row] + (1-p)*optionPrice[now_position - row - 1])
			optionPrice[now_position] = math.Max(keep, earlyExercise)
			now_position++
		}
		row--
	}
}


func stencilRhombus(u, d, p, S, K, r, T float64, N int, depth int, start_position int, loop_i int, loop_j int) {
	var now_position = start_position
	var count_rhombus_block = 0

	var count_rhombus_block_start = N/depth-1 //感覺可以在外面算好 再看看
	for i := 0; i < loop_i; i++ {
		count_rhombus_block += count_rhombus_block_start
		count_rhombus_block_start--
	}
	var left = triangle_block_size * (N/depth + 1) - 1 + rhombus_block_size * (count_rhombus_block + loop_j)
	var right = triangle_block_size * (N/depth) + ((depth-1)*(depth)/2)  + rhombus_block_size * (count_rhombus_block + loop_j + 1)
	
	var row = 0 //從最下一個node的開始
	for i := 0; i < depth; i++ {
		for j := 0; j <= row; j++ {
			
			upper := (loop_j * depth) + (depth - 1) - (row - j)
			down := (N - depth - 2) - (loop_i * depth) - i - upper
			earlyExercise := math.Max(0, S*math.Pow(u, float64(upper))*math.Pow(d, float64(down)) - K)
			if i == 0 && j == 0{
				keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[right] + (1-p)*optionPrice[left])
				optionPrice[now_position] = math.Max(keep, earlyExercise)
				
				now_position++
				left += (depth - 1 + i)
				right += (depth - i)
			}else if j == 0{ //最左邊的 用num的右邊
				keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[now_position - i] + (1-p)*optionPrice[left])
				optionPrice[now_position] = math.Max(keep, earlyExercise)
				now_position++
				left += (depth - 1 - i)

			}else if j == row{ //最右邊的 用num +1 的左邊
				keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[right] + (1-p)*optionPrice[now_position - i - 1])
				optionPrice[now_position] = math.Max(keep, earlyExercise)
				now_position++
				right += (depth - i)

			}else { //中間的 用菱形自己的左右
				keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[now_position-i] + (1-p)*optionPrice[now_position - i - 1])
				optionPrice[now_position] = math.Max(keep, earlyExercise)
				now_position++
			}

		}
		row++
	}

	row = depth - 1
	for i := 0; i < depth - 1; i++ {
		for j := 0; j < row; j++ {
			upper := (loop_j * depth) + j 
			down := (N - 1) - (2 + loop_i)* depth - i -upper -1 // 2是最下的三角形+第二層的菱形 1是因為菱形的下三角式depth上面是depth-1
			earlyExercise := math.Max(0, S*math.Pow(u, float64(upper))*math.Pow(d, float64(down)) - K)
			keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[now_position - row] + (1-p)*optionPrice[now_position - row - 1])
			optionPrice[now_position] = math.Max(keep, earlyExercise)
			now_position++
		}
		row--
	}


}

func americanOptionPrice(S, K, r, q, sigma, T float64, N int, depth int) {
	dt := T / float64(N)
	u := math.Exp((r-q)*dt + sigma*math.Sqrt(dt))
	d := math.Exp((r-q)*dt - sigma*math.Sqrt(dt))
	p := (math.Exp((r-q)*dt) - d) / (u - d)

	var wg sync.WaitGroup
	for i := 0 ; i <= triangle_block_size * ((N / depth) - 1); i += triangle_block_size { // 從最下層往右一個block直到最後
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			stencilTriangle(u, d, p, S, K, r, T, N, depth, i)
		}(i)
	}
	wg.Wait()


	
	var row_count = N / depth //10000 / 25 = 400(三角形有400曾)
	//第一層的菱形(三角形上面的那層)
	for num := 0; num < row_count - 1; num++ { //i 表示現在是第幾個菱形(num)
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			startIndex := (N / depth) * triangle_block_size + num * rhombus_block_size
			stencilSecondLayerRhombus(u, d, p, S, K, r, T, N, depth, startIndex, num)
		}(num)
	}

	wg.Wait()
	
	var row = N / depth - 2
	for i := 0; i < N/depth-2; i++ {
		for j := 0; j < row; j++ {
			wg.Add(1)
			go func(i, j int) {
				defer wg.Done()
				count := count_rhombus(N, depth, i + 1)
				startIndex := (N / depth) * triangle_block_size + (count + j) * rhombus_block_size
				stencilRhombus(u, d, p, S, K, r, T, N, depth, startIndex,i, j)
			}(i, j)
		}
		wg.Wait()
		row--
	}
	
	printMemUsage()
	fmt.Printf("American put option price: %.6f\n", optionPrice[N*(N+1)/2 -1])
}

func main() {
	S := 80.0  // 初始資產價格
	K := 100.0 // 履約價
	r := 0.08  // 無風險利率
	T := 3.0   // 到期時間（年）
	q := 0.12  // 股利率
	sigma := 0.2

	depth := 25



	printMemUsage()
	startTime := time.Now()
	triangle_block_size = (depth * (depth + 1)) / 2
	rhombus_block_size = depth * depth
	
	americanOptionPrice(S, K, r, q, sigma, T, N, depth)
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Printf("Total execution time: %v seconds\n", duration.Seconds())
	
}
