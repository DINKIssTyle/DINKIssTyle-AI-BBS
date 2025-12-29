package main

import (
	"fmt"
	"strings"
)

func main() {
	ua := "Mozilla/3.0 (Macintosh; I; PPC)"
	uaLower := strings.ToLower(ua)

	fmt.Printf("UA: %s\n", ua)
	fmt.Printf("Lower: %s\n", uaLower)

	if strings.Contains(uaLower, "mozilla/2.0") || strings.Contains(uaLower, "mozilla/3.0") ||
		strings.Contains(uaLower, "mozilla/4.0") || strings.Contains(uaLower, "msie") {
		fmt.Println("Detected as DeviceOld")
	} else {
		fmt.Println("Detected as DeviceModern")
	}
}
