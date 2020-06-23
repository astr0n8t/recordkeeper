/*
Package cloudflare is a provider object with support for the Cloudflare API: https://api.cloudflare.com
It exports the required functions for it to be a provider: GetIP and SetIP
It also uses the generic DNS record entry type from the record package
*/
package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/nadehi18/recordkeeper/record"
)

// Cloudflare is the main provider object
type Cloudflare struct {
	username  string
	authToken string
}

type getResponse struct {
	Result   []result `json:"result"`
	Success  bool     `json:"success"`
	Errors   []string `json:"errors"`
	Messages []string `json:"messages"`
	Info     pageInfo `json:"result_info"`
}

type putResponse struct {
	Result   result   `json:"result"`
	Success  bool     `json:"success"`
	Errors   []string `json:"errors"`
	Messages []string `json:"messages"`
	Info     pageInfo `json:"result_info"`
}

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
	entry.Address = address

	// Send the request to update the entry
	putResponse := c.sendPutRequest(entry)

	// Return whether or not we succeeded
	return putResponse.Success
}

// getInfo fills in the gaps from the config file with information from Cloudflare about the entry
func (c *Cloudflare) UpdateEntry(entry *record.Entry) {

	// Check for a zoneID in memory otherwise retrieve it
	if entry.ZoneID == "" {
		entry.ZoneID = c.getZoneID(entry)
	}

	// Pull info from cloudflare
	cloudflareInfo := c.getDomainInfo(entry)

	// Check for a domain ID in memory otherwise retrieve it
	if entry.ID == "" {
		entry.ID = cloudflareInfo.ID
	}

	// Check for a DNS record type in memory otherwise retrieve it
	if entry.RecordType == "" {
		entry.RecordType = cloudflareInfo.RecordType
	}

	// Check for a DNS TTL in memory otherwise retrieve it
	if entry.TTL == 0 {
		entry.TTL = cloudflareInfo.TTL
	}

	// There is no way of knowing if the value in memory is correct or just uninitialized so
	// retrieve the value from Cloudflare
	entry.Proxied = cloudflareInfo.Proxied

	entry.Address = cloudflareInfo.Address
}

// getDomainProperty retrieves a certain data property about the given domain
func (c *Cloudflare) getDomainInfo(entry *record.Entry) result {
	var info result
	// Try to get information about the domain
	domainResponse := c.sendGetRequest(entry, true)

	for _, record := range domainResponse.Result {
		if record.Name == entry.Domain {
			info = record
		}
	}

	// Return the desired property of the domain
	return info
}

// getZoneID tries to get the zone ID of the given domain
func (c *Cloudflare) getZoneID(entry *record.Entry) string {
	var zoneID string
	// Try to get the name of the zone
	zoneName := findZoneName(entry.Domain)
	// Try to get information about the zone
	zoneResponse := c.sendGetRequest(entry, false)

	for _, zone := range zoneResponse.Result {
		if zone.Name == zoneName {
			zoneID = zone.ID
		}
	}

	// Return the zone ID
	return zoneID
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

// sendRequest processes the given arguments to send the appropriate request to the Cloudflare API
func (c *Cloudflare) sendGetRequest(entry *record.Entry, zoneLookup bool) getResponse {
	// The base Cloudflare API URL
	url := "https://api.cloudflare.com/client/v4/zones"

	// If the ID and zoneID are not blank then we are retrieving information about a specific record
	if zoneLookup {
		url += "/" + entry.ZoneID + "/dns_records"
	}

	// Create a new HTTP client and craft the request with the correct headers
	httpClient := http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("X-Auth-Email", c.username)
	request.Header.Add("X-Auth-Key", c.authToken)
	request.Header.Add("X-Content-Type", "application/json")

	if err != nil {
		panic(fmt.Errorf("cannot form CloudFlare API request"))
	}

	// Try to execute the given HTTP request
	resp, err := httpClient.Do(request)
	if err != nil {
		panic(fmt.Errorf("cannot connect to CloudFlare API"))
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(fmt.Errorf("cannot read HTTP response"))
	}

	var responseData getResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		panic(fmt.Errorf("cannot unmarshall JSON from HTTP response"))
	}

	// Return the data from the HTTP request
	return responseData
}

func (c *Cloudflare) sendPutRequest(entry *record.Entry) putResponse {
	// The base Cloudflare API URL
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
	request.Header.Add("X-Auth-Email", c.username)
	request.Header.Add("X-Auth-Key", c.authToken)
	request.Header.Add("X-Content-Type", "application/json")

	if err != nil {
		panic(fmt.Errorf("cannot form CloudFlare API request"))
	}

	// Try to execute the given HTTP request
	resp, err := httpClient.Do(request)
	if err != nil {
		panic(fmt.Errorf("cannot connect to CloudFlare API"))
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(fmt.Errorf("cannot read HTTP response"))
	}

	var responseData putResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		panic(fmt.Errorf("cannot unmarshall JSON from HTTP response"))
	}

	// Return the data from the HTTP request
	return responseData
}
