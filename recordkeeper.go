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
	processConfig()
	providerConnection := providers.GetProvider(viper.GetString("provider"), viper.GetString("username"), viper.GetString("authToken"))
	exitNext := false
	var address string
	refreshInterval := viper.GetInt("interval")
	for !exitNext {
		if viper.GetString("address") == "public" {
			address = publicaddress.GetIP()
		} else {
			address = viper.GetString("address")
		}
		currentAddress := providerConnection.GetIP(viper.GetString("domain"))

		if currentAddress != address {
			changed := providerConnection.SetIP(viper.GetString("domain"), address)

			if changed {
				fmt.Printf("Successfully updated record %v to point to address %v. \n", viper.GetString("domain"), address)
			} else {
				fmt.Printf("ERROR: Unable to change address %v to %v on record %v! \n", currentAddress, address, viper.GetString("domain"))
			}
		}

		if refreshInterval > 0 {
			time.Sleep(time.Duration(refreshInterval) * time.Second)
		} else {
			exitNext = true
		}
	}
}

func processConfig() {

	viper.SetDefault("provider", "cloudflare")
	viper.SetDefault("interval", 60)
	viper.SetDefault("address", "public")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/recordkeeper/")
	viper.AddConfigPath("$HOME/.config/recordkeeper")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	pflag.String("provider", "cloudflare", "Selects a DNS provider to use")
	pflag.String("username", "", "The username to use to connect to the DNS service")
	pflag.String("authToken", "", "The authentication token to connect to the DNS provider")
	pflag.String("domain", "", "The domain record to check")
	pflag.String("address", "public", "The address to bind the domain to (use public to specify current public IP)")
	pflag.Int("interval", 60, "The time in seconds to check the DNS record, set to 0 to only run once")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	if viper.GetString("username") == "" || viper.GetString("authToken") == "" || viper.GetString("domain") == "" {
		panic(fmt.Errorf("one or more required arguments not supplied or config file could not be read\n required arguments: username, authToken, domain"))
	}

}
