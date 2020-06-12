package main

import (
	"fmt"
	"time"

	"github.com/nadehi18/recordkeeper/providers"
)

func main() {

	providerConnection := providers.GetProvider(cfg.providerName, cfg.username, cfg.authToken)
	exitNext := false
	for !exitNext {
		currentAddress := providerConnection.GetIP(cfg.domain)

		if currentAddress != cfg.address {
			changed := providerConnection.SetIP(cfg.address)

			if changed {
				fmt.Printf("Successfully updated record %v to point to address %v. \n", cfg.domain, cfg.address)
			} else {
				fmt.Printf("ERROR: Unable to change address %v to %v on record %v! \n", currentAddress, cfg.address, cfg.domain)
			}
		}
		time.Sleep(cfg.interval * time.Second)
	}
}
