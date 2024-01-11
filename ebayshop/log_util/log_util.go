package log_util

import "fmt"

func PrintFlag(serviceName string, debug bool, message string) {
	if debug {
		fmt.Printf("[ServiceName=%v], Log: %v\n", serviceName, message)
	}
}
