/*
Package publicaddress has one function which gets the current public IP address
of whatever machine it is run on.  It uses the ipify.org API: https://ipify.org
*/
package publicaddress

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetIP returns the current public IP address
func GetIP() string {
	var ip string

	resp, err := http.Get("https://api.ipify.org")

	if err != nil {
		panic(fmt.Errorf("cannot retrieve public IP address"))
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(fmt.Errorf("cannot retrieve public IP address"))
		}
		ip = string(body)
	}

	return ip
}
