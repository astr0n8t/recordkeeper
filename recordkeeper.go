package main

import (
	"fmt"
	"time"

	"github.com/nadehi18/recordkeeper/providers"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	processConfig()
	providerConnection := providers.GetProvider(viper.GetString("provider"), viper.GetString("username"), viper.GetString("authToken"))
	exitNext := false
	for !exitNext {
		currentAddress := providerConnection.GetIP(viper.GetString("domain"))

		if currentAddress != viper.GetString("address") {
			changed := providerConnection.SetIP(viper.GetString("address"))

			if changed {
				fmt.Printf("Successfully updated record %v to point to address %v. \n", viper.GetString("domain"), viper.GetString("address"))
			} else {
				fmt.Printf("ERROR: Unable to change address %v to %v on record %v! \n", currentAddress, viper.GetString("address"), viper.GetString("domain"))
			}
		}
		time.Sleep(time.Duration(viper.GetInt("interval")) * time.Second)
	}
}

func processConfig() {

	viper.SetDefault("provider", "cloudflare")
	viper.SetDefault("interval", 60)

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/recordkeeper/")
	viper.AddConfigPath("$HOME/.config/recordkeeper")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	pflag.String("provider", "cloudflare", "Selects a DNS provider to use")
	pflag.String("username", "null", "The username to use to connect to the DNS service")
	pflag.String("authToken", "null", "The authentication token to connect to the DNS provider")
	pflag.String("domain", "null", "The domain record to check")
	pflag.String("address", "public", "The address to bind the domain to (use public to specify current public IP)")
	pflag.Int("interval", 60, "The time in seconds to check the DNS record")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

}
