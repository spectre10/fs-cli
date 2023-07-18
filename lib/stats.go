package lib

import (
	"fmt"
	"time"
)

// Start returns the starting time of the transaction in milliseconds
func Start() int64 {
	return time.Now().UnixMilli()
}

// FinalStat prints the final stats of the transaction
func FinalStat(fileSize uint64, startTime int64) {
	currentTime := time.Now().UnixMilli()
	timeDiff := float64(currentTime-startTime) / 1000.0
	fileSizeInMiB := float64(fileSize) / 1048576.0

	fmt.Printf("\nStats:\n")
	fmt.Printf("Time Taken: %.2f seconds\n", timeDiff)
	fmt.Printf("Total Amount Transfered: %.2f MiB\n", fileSizeInMiB)
	fmt.Printf("Average Speed: %.2f MiB/s\n", fileSizeInMiB/timeDiff)
}
