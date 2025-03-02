package main
///可輸入length來調整 每個block的大小
import (
	"fmt"
	"math"
	"time"
	"sync"
	"github.com/viney-shih/goroutines"
)

const sigma = 0.2



func americanOptionPrice(u, d, p, S, K, r, T float64, N int, depth int ,pool_num int) float64 {
	

	optionTree := make([][]float64, N+1)
	for i := range optionTree {
		optionTree[i] = make([]float64, i+1)
		for j := range optionTree[i] {
			optionTree[i][j] = -1.0
		}
	}
	pool := goroutines.NewPool( 
		pool_num, //pool的最大大小 最多有幾個worker可以同時運行(控制最大併發數量)
		//goroutines.WithTaskQueueLength(20), //queue中可以容納的task數量 當現在滿就要等到有位置才能進來(控制排隊數量)
		//goroutines.WithPreAllocWorkers(1), //預先配置worker(控制是否要預先配置worker)
	)
	defer pool.Release()


	// 初始化只有最後一層的選擇權價格
	for j := 0; j < N; j++ {
		optionTree[N-1][j] = math.Max(0, S*math.Pow(u, float64(j))*math.Pow(d, float64(N-1-j))-K)
	}

	count := 0
	last := N

	var wg sync.WaitGroup

	for i := N - depth; i >= 0; i -= depth { //從總層數-深度開始，每次減少深度 eg. 10000-1000 ,9000-8000 ...
		wg.Add(i + 1)
		for j := 0; j <= i; j++ {
			pool.Schedule(func() {
				defer wg.Done()
				treeLengthAdjustment := 0
				for m := N - (count * depth) - 1; m >= i; m-- {
					for tree_length := depth - 1 - treeLengthAdjustment; tree_length >= 0; tree_length-- {
						if optionTree[m][j+tree_length] >= 0.0 {
							continue
						}
						earlyExercise := math.Max(0, S*math.Pow(u, float64(j+tree_length))*math.Pow(d, float64(m-j-tree_length)))
						keep := 0.0
						if m != N - 1 {
							keep = math.Exp(-r*(T/float64(N))) * (p*optionTree[m+1][j+tree_length+1] + (1-p)*optionTree[m+1][j+tree_length])
						
						}
						optionTree[m][j+tree_length] = math.Max(keep, earlyExercise-K)
					}
					treeLengthAdjustment++
					}
				})
		}
		wg.Wait()
		count++
		last = i
	}

	if optionTree[0][0] < 0.0 {
		for i := last - 1; i >= 0; i-- {
			for j := 0; j <= i; j++ {
				earlyExercise := math.Max(S*math.Pow(u, float64(j)) * math.Pow(d, float64(i-j)), 0 )
				keep := math.Exp(-r*(T/float64(N))) * (p*optionTree[i+1][j+1] + (1-p)*optionTree[i+1][j])
				optionTree[i][j] = math.Max(keep, earlyExercise-K)
			}
		}
	}

	return optionTree[0][0]
}


func main() {
	S := 80.0  // 初始資產價格
	K := 100.0 // 履約價
	r := 0.08  // 無風險利率
	T := 3.0   // 到期時間（年）
	N := 10000  // 樹層數
	q := 0.12  // 股利率
	pool_size := 500

	var one_length int
	fmt.Printf("Please input the number of tree depth: ")
	fmt.Scanln(&one_length)

	startTime := time.Now()
	dt := T / float64(N)
	u := math.Exp((r-q)*dt + sigma*math.Sqrt(dt))
	d := math.Exp((r-q)*dt - sigma*math.Sqrt(dt))
	p := (math.Exp((r-q)*dt) - d) / (u - d)
	optionPrice := americanOptionPrice(u, d, p, S, K, r, T, N, one_length, pool_size)
	fmt.Printf("American Option Price: %v\n", optionPrice)
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Printf("Total execution time: %v seconds\n", duration.Seconds())
}