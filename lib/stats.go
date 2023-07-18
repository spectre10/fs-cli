package lib

import (
	"fmt"
	"time"

	"github.com/vbauerster/mpb/v8/decor"
)

// Start returns the starting time of the transaction in milliseconds
func Start() int64 {
	return time.Now().UnixMilli()
}

// FinalStat prints the final stats of the transaction
func FinalStat(TotalAmount uint64, startTime int64) {
	currentTime := time.Now().UnixMilli()
	timeDiff := float64(currentTime-startTime) / 1000.0
	totalAmountInMiB := float64(TotalAmount) / 1048576.0

	fmt.Printf("\nStats:\n")
	fmt.Printf("Time Taken: %.2f seconds\n", timeDiff)
	fmt.Printf("Total Amount Transfered: % .2f \n", decor.SizeB1024(TotalAmount))
	fmt.Printf("Average Speed: %.2f MiB/s\n", totalAmountInMiB/timeDiff)
}
