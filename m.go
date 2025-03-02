package main
//M字存法 省記憶體
import (
	"fmt"
	"math"
	"time"
	"sync"
	"runtime"
)

var optionPrice []float64
var last_layer_num []float64
var last_layer_num_temporary []float64
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

func stencilTriangle(u, d, p, S, K, r, T float64, N int, depth int, start_position int, num_of_tri int) {
	var now_position int = start_position
	
	keep_left := depth + (2*depth)*(num_of_tri - 1)
	keep_right:= keep_left + depth
	if num_of_tri == 0{
		keep_right = 0
	}


	for i := 0; i < depth; i++ {	
		upper := now_position % triangle_block_size + (start_position / triangle_block_size) * depth //原本是start是now有問題再改回來		
		down := N - 1 - upper
		optionPrice[now_position] = math.Max(0, S * math.Pow(u, float64(upper) ) * math.Pow(d, float64(down)) - K)
		if num_of_tri == 0 && i == depth - 1{
			last_layer_num[keep_right] = optionPrice[now_position]
			keep_right++
		}else if num_of_tri != 0 && i == 0 {
			last_layer_num[keep_left] = optionPrice[now_position]
			keep_left++
		}else if num_of_tri != N/depth-1 && i == depth-1{
			last_layer_num[keep_right] = optionPrice[now_position]
			keep_right++
		}

		now_position++
	}

	var row int = depth - 1
	for i := 0; i < depth - 1; i++ { 
		for j := 0; j < row; j++ {
			var upper = (start_position / triangle_block_size)*depth + j
			var down = (N - 1) - (i+1) - upper

			earlyExercise := math.Max(0, S*math.Pow(u, float64(upper))*math.Pow(d, float64(down)) - K)
			keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[now_position - row] + (1-p)*optionPrice[now_position - row - 1])
			optionPrice[now_position] = math.Max(keep, earlyExercise)

			if num_of_tri != 0 && j == 0 && j == row - 1{ //最上面的 左右都存菱形比較好算
				last_layer_num[keep_left] = optionPrice[now_position]
				last_layer_num[keep_right] = optionPrice[now_position]
				
			}else if num_of_tri != (N / depth-1) && j == row - 1{ //右邊的
				last_layer_num[keep_right] = optionPrice[now_position]
				keep_right++
			}else if num_of_tri != 0 && j == 0{ //左邊的
				last_layer_num[keep_left] = optionPrice[now_position]
				keep_left++
			}
			now_position++
		}
		row--
	}

}

func stencilRhombus(u, d, p, S, K, r, T float64, N int, depth int, start_position int, loop_i int, num int, outer_row int, ) {
	var now_position = start_position

	var left = num * (2 * depth)
	var right = left + depth

	var last_layer_left = depth + (num-1) * (depth*2)
	var last_layer_right = last_layer_left + depth
	if num == 0{
		last_layer_right = 0
	}

	
	var row = 0 //從最下一個node的開始
	for i := 0; i < depth; i++ {
		for j := 0; j <= row; j++ {
			
			upper := (num * depth) + (depth - 1) - (row - j)
			down := (N - 2) - (loop_i * depth) - i - upper
			earlyExercise := math.Max(0, S*math.Pow(u, float64(upper))*math.Pow(d, float64(down)) - K)
			
			if i == 0 && j == 0{
				keep := math.Exp(-r*(T/float64(N))) * (p*last_layer_num[right] + (1-p)*last_layer_num[left])
				optionPrice[now_position] = math.Max(keep, earlyExercise)
				now_position++
				left ++
				right ++
			}else if j == 0{ //最左邊的 用num的右邊
				keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[now_position - i] + (1-p)*last_layer_num[left])
				optionPrice[now_position] = math.Max(keep, earlyExercise)
				now_position++
				left++

				if num != 0 && i == depth - 1 { //倒三角形最上面那排的左邊 (0的左邊不用存)
					last_layer_num_temporary[last_layer_left] = optionPrice[now_position-1]
					last_layer_left++
				}

			}else if j == row{ //最右邊的 用num +1 的左邊
				keep := math.Exp(-r*(T/float64(N))) * (p*last_layer_num[right] + (1-p)*optionPrice[now_position - i - 1])
				optionPrice[now_position] = math.Max(keep, earlyExercise)
				now_position++
				right++

				if i == depth - 1 && outer_row != num{ //倒三角形最上面那排的右邊 最右邊的不用算
					last_layer_num_temporary[last_layer_right] = optionPrice[now_position-1]
					last_layer_right++
				}

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
			upper := (num * depth) + j 
			down := (N - 1) - (1 + loop_i)* depth - i -upper - 1 // 1是最下的三角形 1是因為菱形的下三角式depth上面是depth-1
			
			earlyExercise := math.Max(0, S*math.Pow(u, float64(upper))*math.Pow(d, float64(down)) - K)
			keep := math.Exp(-r*(T/float64(N))) * (p*optionPrice[now_position - row] + (1-p)*optionPrice[now_position - row - 1])
			optionPrice[now_position] = math.Max(keep, earlyExercise)


			if j==0 && j == row-1 && num != 0 && outer_row != num{ 
				last_layer_num_temporary[last_layer_left] = optionPrice[now_position]
				last_layer_num_temporary[last_layer_right] = optionPrice[now_position]
				
			}else if num != 0 && j == 0{ //左邊 0的左邊不用存
				last_layer_num_temporary[last_layer_left] = optionPrice[now_position]
				last_layer_left++
			}else if outer_row != num && j == row - 1{ //右邊 且不是這整行最右邊的
				last_layer_num_temporary[last_layer_right] = optionPrice[now_position]
				last_layer_right++
			}

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

	startTime := time.Now()
	var wg sync.WaitGroup
	for i := 0 ; i <= triangle_block_size * ((N / depth) - 1); i += triangle_block_size  {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			stencilTriangle(u, d, p, S, K, r, T, N, depth, i, i / triangle_block_size )
		}(i)
	}
	wg.Wait()

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Printf("Triangle time: %v seconds\n", duration.Seconds())

	row := N / depth - 1 //從倒數第二層開始
	for i := 0; i < N/depth - 1; i++ {
		for j := 0; j < row; j++ { //i當作菱形的num
			wg.Add(1)
			go func(i, j int) {
				defer wg.Done()
				startIndex := j * rhombus_block_size
				stencilRhombus(u, d, p, S, K, r, T, N, depth, startIndex,i, j, row-1)
		
			}(i, j)
		}
		wg.Wait()
		for m := 0; m < (row-1)*depth*2; m++{
			last_layer_num[m] = last_layer_num_temporary[m]
		}
		row--
	}
	printMemUsage() 


}

func main() {
	S := 80.0  // 初始資產價格
	K := 100.0 // 履約價
	r := 0.08  // 無風險利率
	T := 3.0   // 到期時間（年）
	N := 10000 // 樹層數
	q := 0.12  // 股利率
	sigma := 0.2

	depth := 25


	fmt.Printf("Depth: %v\n", depth)

	
	startTime := time.Now()

	triangle_block_size = (depth * (depth + 1)) / 2
	rhombus_block_size = depth * depth
	optionPrice = make([]float64, (N / depth -1)* rhombus_block_size) //開最大的大小是第二層菱形的大小
	last_layer_num = make([]float64, (N /depth) * 2 * depth) //用來存M字的那邊數字
	last_layer_num_temporary = make([]float64, (N /depth) * 2 * depth) //用來暫存菱形的M字
	printMemUsage() 
	
	
	
	
	americanOptionPrice(S, K, r, q, sigma, T, N, depth)
	endTime := time.Now()
	duration := endTime.Sub(startTime)


	fmt.Printf("American put option price: %.6f\n", optionPrice[rhombus_block_size-1])
	fmt.Printf("Total execution time: %v seconds\n", duration.Seconds())
	
}
