package main

import (
	"fmt"
	"os"

	"github.com/beevik/ntp"
)

func main() {
	ntpTime, err := ntp.Time("ntp0.ntp-servers.net")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Println(ntpTime)
}
