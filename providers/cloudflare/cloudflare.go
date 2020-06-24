/*
Package cloudflare is a provider object with support for the Cloudflare API: https://api.cloudflare.com
It exports the required functions for it to be a provider: UpdateEntry and SetIP
It also uses the generic DNS record entry type from the record package
*/
package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/nadehi18/recordkeeper/record"
)

// Cloudflare is the main provider object
type Cloudflare struct {
	username  string
	authToken string
}

// Struct to store the response from a GET request
type getResponse struct {
	Result   []result `json:"result"`
	Success  bool     `json:"success"`
	Errors   []string `json:"errors"`
	Messages []string `json:"messages"`
	Info     pageInfo `json:"result_info"`
}

// Struct to store the response from a PUT request
type putResponse struct {
	Result   result   `json:"result"`
	Success  bool     `json:"success"`
	Errors   []string `json:"errors"`
	Messages []string `json:"messages"`
	Info     pageInfo `json:"result_info"`
}

// Struct to store the information in the result of a response
type result struct {
	ID         string `json:"id"`
	ZoneID     string `json:"zone_id"`
	Name       string `json:"name"`
	RecordType string `json:"type"`
	Address    string `json:"content"`
	Proxiable  bool   `json:"proxiable"`
	Proxied    bool   `json:"proxied"`
	TTL        int    `json:"ttl"`
}

// Struct to store the information in the info of a response
type pageInfo struct {
	CurrentPage int `json:"page"`
	PerPage     int `json:"per_page"`
	Count       int `json:"count"`
	TotalCount  int `json:"total_count"`
	TotalPages  int `json:"total_pages"`
}

// New returns a pointer to a new initialized Cloudflare object
func New(user string, auth string) *Cloudflare {
	// Return the reference to a new Cloudflare object
	return &Cloudflare{user, auth}
}

// SetIP attempts to change the IP address stored in Cloudflare's database for the current entry to
// the value stored in address.  It then returns whether or not the attempt was successful.
func (c *Cloudflare) SetIP(address string, entry *record.Entry) bool {
	// Get the information about the entry
	c.UpdateEntry(entry)
	// Set the new address
	entry.Address = address

	// Send the request to update the entry
	putResponse := c.sendPutRequest(entry)

	// Return whether or not we succeeded
	return putResponse.Success
}

// UpdateEntry updates the entry with values from Cloudflare
func (c *Cloudflare) UpdateEntry(entry *record.Entry) {

	// Check for a zoneID in memory otherwise retrieve it
	if entry.ZoneID == "" {
		c.getZoneID(entry)
	}

	// Pull info from cloudflare
	cloudflareInfo := c.getDomainInfo(entry)

	// Check for a domain ID in memory otherwise retrieve it
	if entry.ID == "" {
		entry.ID = cloudflareInfo.ID
	}

	// Update values in memory
	entry.RecordType = cloudflareInfo.RecordType
	entry.TTL = cloudflareInfo.TTL
	entry.Proxied = cloudflareInfo.Proxied
	entry.Address = cloudflareInfo.Address
}

// getDomainInfo retrieves information about the given domain
func (c *Cloudflare) getDomainInfo(entry *record.Entry) result {
	var info result
	found := false

	currentPage := 1

	// Find the correct record in the response
	// Choose from ID if it is available, otherwise choose based on name
	for !found {
		// Try to get information about the domain
		domainResponse := c.sendGetRequest(entry, true, currentPage)
		for _, record := range domainResponse.Result {
			if entry.ID == record.ID || (entry.ID == "" && record.Name == entry.Domain) {
				info = record
				found = true
			}
		}
		// Check if there are more pages if we have not found the record
		if !found {
			if domainResponse.Info.CurrentPage < domainResponse.Info.TotalPages {
				currentPage++
			} else {
				panic(fmt.Errorf("cannot retrieve information for given record"))
			}
		}
	}

	// Return the info on the domain
	return info
}

// getZoneID tries to get the zone ID of the given domain
func (c *Cloudflare) getZoneID(entry *record.Entry) {
	found := false
	currentPage := 1
	// Try to get the name of the zone
	zoneName := findZoneName(entry.Domain)
	for !found {
		// Try to get information about the zone
		zoneResponse := c.sendGetRequest(entry, false, currentPage)

		// Select the correct zoneID based on the zone name
		for _, zone := range zoneResponse.Result {
			if zone.Name == zoneName {
				entry.ZoneID = zone.ID
				found = true
			}
		}
		// Check if there are more pages if we have not found the zone
		if !found {
			if zoneResponse.Info.CurrentPage < zoneResponse.Info.TotalPages {
				currentPage++
			} else {
				panic(fmt.Errorf("cannot retrieve information for zone of given record"))
			}
		}
	}
}

// findZoneName tries to get the name of the zone that the given domain is in by splitting
// the string and assuming that the zone name is in the domain name
func findZoneName(domain string) string {
	// Split the domain string into a slice
	zoneNameSlice := strings.Split(domain, ".")

	var zoneName string
	// Check that the domain is actually more than one part to avoid an out of bounds exception
	if len(zoneNameSlice) > 1 {
		zoneName = zoneNameSlice[len(zoneNameSlice)-2] + "." + zoneNameSlice[len(zoneNameSlice)-1]
	} else {
		// Otherwise just set the zone name as the original domain name
		zoneName = domain
	}

	// Return the supposed zone name
	return zoneName
}

// sendGetRequest processes the given arguments to send the appropriate GET request to the Cloudflare API
func (c *Cloudflare) sendGetRequest(entry *record.Entry, zoneLookup bool, pageNumber int) getResponse {
	// The base Cloudflare API URL
	url := "https://api.cloudflare.com/client/v4/zones"

	// If the ID and zoneID are not blank then we are retrieving information about a specific record
	if zoneLookup {
		url += "/" + entry.ZoneID + "/dns_records"
	}

	// Add current page number to request
	url += "?page=" + strconv.Itoa(pageNumber)

	// Create a new HTTP client and craft the request with the correct headers
	httpClient := http.Client{}
	request, err := http.NewRequest("GET", url, nil)

	// Check if user is using user service key instead of API key
	if c.username == "USERSERVICEKEY" {
		request.Header.Add("X-Auth-User-Service-Key", c.authToken)
	} else {
		request.Header.Add("X-Auth-Email", c.username)
		request.Header.Add("X-Auth-Key", c.authToken)
	}
	request.Header.Add("X-Content-Type", "application/json")

	if err != nil {
		panic(fmt.Errorf("cannot form CloudFlare API request"))
	}

	// Try to execute the given HTTP request
	resp, err := httpClient.Do(request)
	if err != nil {
		panic(fmt.Errorf("cannot connect to CloudFlare API"))
	}

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(fmt.Errorf("cannot read HTTP response"))
	}
	// Marshal the JSON response into a get response struct type
	var responseData getResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		panic(fmt.Errorf("cannot unmarshall JSON from HTTP response"))
	}

	// Return the data from the HTTP request
	return responseData
}

// sendPutRequest processes the given arguments to send the appropriate PUT request to the Cloudflare API
func (c *Cloudflare) sendPutRequest(entry *record.Entry) putResponse {
	// The url of the correct API call
	url := "https://api.cloudflare.com/client/v4/zones" + "/" + entry.ZoneID + "/dns_records/" + entry.ID

	var data []byte
	// Attempt to marshall the given record entry informatino into JSON format
	marshalledData, err := json.Marshal(entry)
	if err != nil {
		panic(fmt.Errorf("cannot marshall json data for domain %v", entry.Domain))
	} else {
		// Store the marshalled JSON data
		data = marshalledData
	}

	// Create a new HTTP client and craft the request with the correct headers
	httpClient := http.Client{}
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))

	// Check if user is using user service key instead of API key
	if c.username == "USERSERVICEKEY" {
		request.Header.Add("X-Auth-User-Service-Key", c.authToken)
	} else {
		request.Header.Add("X-Auth-Email", c.username)
		request.Header.Add("X-Auth-Key", c.authToken)
	}
	request.Header.Add("X-Content-Type", "application/json")

	if err != nil {
		panic(fmt.Errorf("cannot form CloudFlare API request"))
	}

	// Try to execute the given HTTP request
	resp, err := httpClient.Do(request)
	if err != nil {
		panic(fmt.Errorf("cannot connect to CloudFlare API"))
	}

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(fmt.Errorf("cannot read HTTP response"))
	}
	// Marshal the JSON response into a put response struct type
	var responseData putResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		panic(fmt.Errorf("cannot unmarshall JSON from HTTP response"))
	}

	// Return the data from the HTTP request
	return responseData
}
