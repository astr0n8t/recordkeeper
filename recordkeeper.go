/*
----------------RecordKeeper----------------
A dynamic DNS client written in Go with
support for tracking the current public
IP or a pre-configured address.
------------Providers Supported-------------
Cloudflare
--------------------------------------------
Written by Nathan Higley (@nadehi18)
contact@nathanhigley.com
https://nathanhigley.com
https://github.com/nadehi18
--------------------------------------------
*/

package main

import (
	"fmt"
	"time"

	"github.com/nadehi18/recordkeeper/providers"
	"github.com/nadehi18/recordkeeper/publicaddress"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	// Reads in the configuration file and/or arguments
	processConfig()
	// Gets the respective struct for the selected provider
	providerConnection := providers.GetProvider(viper.GetString("provider"), viper.GetString("username"), viper.GetString("authToken"))
	// Controls the event loop
	exitNext := false
	// Stores the IP to set to, probably will be changed to an array
	var address string
	// Stores the time in minutes to refresh the records
	refreshInterval := viper.GetInt("interval")

	// The main event loop
	for !exitNext {
		// Checks whether the address is the current public address or not
		if viper.GetString("address") == "public" {
			// Get the current public address
			address = publicaddress.GetIP()
		} else {
			// Get the preset address in the config
			address = viper.GetString("address")
		}
		// Get the current record address from the DNS provider
		currentAddress := providerConnection.GetIP(viper.GetString("domain"))

		// Check if the address in the record differs from the user set address
		if currentAddress != address {
			// If it has, change the record to point to the user set address
			changed := providerConnection.SetIP(viper.GetString("domain"), address)

			// Check if the record was successfully changed
			if changed {
				fmt.Printf("Successfully updated record %v to point to address %v. \n", viper.GetString("domain"), address)
			} else {
				fmt.Printf("ERROR: Unable to change address %v to %v on record %v! \n", currentAddress, address, viper.GetString("domain"))
			}
		} else {
			fmt.Printf("Record %v still points at address %v.  No errors encountered.\n", viper.GetString("domain"), currentAddress)
		}

		// Check if the loop should continue or exit
		// A refresh interval of 0 or lower indicates no
		if refreshInterval > 0 {
			// Sleep the number of minutes
			time.Sleep(time.Duration(refreshInterval) * time.Minute)
		} else {
			// Exit the loop
			exitNext = true
		}
	}
}

// Processes the configuration file and command line arguments using Viper and PFlags
func processConfig() {

	// Set defaults
	viper.SetDefault("provider", "cloudflare")
	viper.SetDefault("interval", 60)
	viper.SetDefault("address", "public")

	// Set default config directories
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/recordkeeper/")
	viper.AddConfigPath("$HOME/.config/recordkeeper")
	viper.AddConfigPath(".")

	// Read the config
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Config file not found, reading command line arguments only.\n")
		} else {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}

	// Add and read command line arguments
	pflag.String("provider", "cloudflare", "Selects a DNS provider to use")
	pflag.String("username", "", "The username to use to connect to the DNS service")
	pflag.String("authToken", "", "The authentication token to connect to the DNS provider")
	pflag.String("domain", "", "The domain record to check")
	pflag.String("address", "public", "The address to bind the domain to (use public to specify current public IP)")
	pflag.Int("interval", 60, "The time in seconds to check the DNS record, set to 0 to only run once")
	pflag.Parse()

	// Add the command line arguments to viper
	viper.BindPFlags(pflag.CommandLine)

	// Check for required configuration options
	if viper.GetString("username") == "" || viper.GetString("authToken") == "" || viper.GetString("domain") == "" {
		panic(fmt.Errorf("one or more required arguments not supplied or config file could not be read\n required arguments: username, authToken, domain"))
	}
}
