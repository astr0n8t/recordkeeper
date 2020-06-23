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
	"github.com/nadehi18/recordkeeper/record"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	// Reads in the configuration file and/or arguments
	options := processConfig()
	// Gets the respective struct for the selected provider
	providerConnection := providers.GetProvider(options.Provider, options.Username, options.AuthToken)
	// Controls the event loop
	exitNext := false
	// Stores the IP to set to, probably will be changed to an array
	var address string
	// Stores the time in minutes to refresh the records
	refreshInterval := options.Interval
	// The main event loop
	for !exitNext {
		// Iterate for every DNS entry defined in the config
		for _, entry := range options.Entries {
			// Checks whether the address is the current public address or not
			if entry.Address == "public" {
				// Get the current public address
				address = publicaddress.GetIP()
			} else {
				// Get the preset address in the config
				address = entry.Address
			}
			// Get the current record address from the DNS provider
			currentAddress := providerConnection.GetIP(entry)

			// Check if the address in the record differs from the user set address
			if currentAddress != address {
				// If it has, change the record to point to the user set address
				changed := providerConnection.SetIP(address, entry)

				// Check if the record was successfully changed
				if changed {
					fmt.Printf("Successfully updated record %v to point to address %v. \n", entry.Domain, address)
				} else {
					fmt.Printf("ERROR: Unable to change address %v to %v on record %v! \n", currentAddress, address, entry.Domain)
				}
			} else {
				fmt.Printf("Record %v still points at address %v.  No errors encountered.\n", entry.Domain, currentAddress)
			}
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

// A struct to store configuration options
type config struct {
	Provider  string         `mapstructure:"provider"`
	Username  string         `mapstructure:"username"`
	AuthToken string         `mapstructure:"authToken"`
	Interval  int            `mapstructure:"interval"`
	Entries   []record.Entry `mapstructure:"records"`
}

// Processes the configuration file and command line arguments using Viper and PFlags
func processConfig() config {

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
	pflag.Int("interval", 60, "The time in seconds to check the DNS record, set to 0 to only run once")
	pflag.Parse()

	// Add the command line arguments to viper
	viper.BindPFlags(pflag.CommandLine)

	// Unmarshall the config file into the config struct
	var processedConfig config
	err = viper.Unmarshal(&processedConfig)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshall config file or command line arguments"))
	}

	// Check for required configuration options
	if processedConfig.Username == "" || processedConfig.AuthToken == "" || processedConfig.Entries[0] == (record.Entry{}) || processedConfig.Entries[0].Domain == "" {
		panic(fmt.Errorf("one or more required arguments not supplied or config file could not be read\n required arguments: username, authToken, domain"))
	}

	// Return the config struct with config data
	return processedConfig
}
