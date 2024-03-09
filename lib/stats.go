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

func GetStats(TotalAmount uint64, startTime int64) (float64, decor.SizeB1024, float64) {
	currentTime := time.Now().UnixMilli()
	timeTaken := float64(currentTime-startTime) / 1000.0
	totalAmountInMiB := float64(TotalAmount) / 1048576.0
	speed := totalAmountInMiB / timeTaken
	return timeTaken, decor.SizeB1024(TotalAmount), speed
}

// FinalStat prints the final stats of the transaction
func FinalStat(totalAmount uint64, startTime int64) {
	timeTaken, amount, speed := GetStats(totalAmount, startTime)
	fmt.Printf("\nStats:\n")
	fmt.Printf("Time Taken: %.2f seconds\n", timeTaken)
	fmt.Printf("Total Amount Transferred: % .2f \n", amount)
	fmt.Printf("Average Speed: %.2f MiB/s\n", speed)
}
