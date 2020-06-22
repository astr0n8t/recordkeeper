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
	"net/http"
	"strconv"
	"strings"

	"github.com/nadehi18/recordkeeper/record"
)

// Cloudflare is the main provider object
type Cloudflare struct {
	username  string
	authToken string
	records   map[string]*record.Entry
}

// New returns a pointer to a new initialized Cloudflare object
func New(user string, auth string) *Cloudflare {
	// Return the reference to a new Cloudflare object
	return &Cloudflare{user, auth, make(map[string]*record.Entry)}
}

// GetIP returns the IP address stored in Cloudflare's database for the current entry
func (c *Cloudflare) GetIP(entry record.Entry) string {
	// Get the information about the entry
	c.getInfo(entry)

	// Send a request to get the IP address of the entry stored in Cloudflare's database
	response := c.sendRequest(c.records[entry.Domain].ZoneID, c.records[entry.Domain].ID, entry.Domain, "GET")

	// Decode the data into a string interface map
	var address string
	var recordData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&recordData)
	// Make sure the request was successful and retrieve the IP address
	if itemToString(recordData["success"]) == "true" {
		recordDetails := recordData["result"].(map[string]interface{})
		address = itemToString(recordDetails["content"])
	} else {
		panic(fmt.Errorf("cannot get CloudFlare API response for record %v", entry.Domain))
	}

	// Return the address
	return address
}

// SetIP attempts to change the IP address stored in Cloudflare's database for the current entry to
// the value stored in address.  It then returns whether or not the attempt was successful.
func (c *Cloudflare) SetIP(address string, entry record.Entry) bool {
	// Get the information about the entry
	c.getInfo(entry)

	// Set the address in memory
	c.records[entry.Domain].Address = address
	// Send the request to update the entry
	response := c.sendRequest(c.records[entry.Domain].ZoneID, c.records[entry.Domain].ID, entry.Domain, "PUT")

	// Check if the request was successful
	success := false
	var changeData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&changeData)
	if itemToString(changeData["success"]) == "true" {
		success = true
	}

	// Return whether or not we succeeded
	return success
}

// getInfo fills in the gaps from the config file with information from Cloudflare about the entry
func (c *Cloudflare) getInfo(entry record.Entry) {
	// Check if the record is already in the Cloudflare object records
	// if it is not then copy the data from the given entry, otherwise continue
	_, exists := c.records[entry.Domain]
	if !exists {
		c.records[entry.Domain] = &entry
	}

	// Check for a zoneID in memory otherwise retrieve it
	if c.records[entry.Domain].ZoneID == "" {
		c.records[entry.Domain].ZoneID = c.getZoneID(entry.Domain)
	}

	// Check for a domain ID in memory otherwise retrieve it
	if c.records[entry.Domain].ID == "" {
		c.records[entry.Domain].ID = c.getDomainProperty(entry.Domain, "id")
	}

	// Check for a DNS record type in memory otherwise retrieve it
	if c.records[entry.Domain].RecordType == "" {
		c.records[entry.Domain].RecordType = c.getDomainProperty(entry.Domain, "type")
	}

	// Check for a DNS TTL in memory otherwise retrieve it
	if c.records[entry.Domain].TTL == 0 {
		ttl, err := strconv.Atoi(c.getDomainProperty(entry.Domain, "ttl"))
		if err == nil {
			c.records[entry.Domain].TTL = ttl
		}
	}

	// There is no way of knowing if the value in memory is correct or just uninitialized so
	// retrieve the value from Cloudflare
	proxiedS := c.getDomainProperty(entry.Domain, "proxied")
	if proxiedS == "true" {
		c.records[entry.Domain].Proxied = true
	} else {
		c.records[entry.Domain].Proxied = false
	}
}

// getDomainProperty retrieves a certain data property about the given domain
func (c *Cloudflare) getDomainProperty(domain string, property string) string {
	// Try to get information about the domain
	response := c.sendRequest(c.records[domain].ZoneID, "", domain, "GET")

	// Decode the data into a slice of interfaces
	var value string
	var domainData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&domainData)
	domains := domainData["result"].([]interface{})
	// Iterate over the slices
	for i := range domains {
		// Check if this is the correct domain
		currentDomain := domains[i].(map[string]interface{})
		currentDomainName := itemToString(currentDomain["name"])
		if currentDomainName == domain {
			// Get the desired property of the domain
			value = itemToString(currentDomain[property])

		}
	}

	// Return the desired property of the domain
	return value
}

// getZoneID tries to get the zone ID of the given domain
func (c *Cloudflare) getZoneID(domain string) string {
	var zoneID string
	// Try to get the name of the zone
	zoneName := findZoneName(domain)
	// Try to get information about the zone
	response := c.sendRequest("", "", domain, "GET")

	// Decode the data into a slice of interfaces
	var zoneData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&zoneData)
	zones := zoneData["result"].([]interface{})
	// Iterate over the slices
	for i := range zones {
		// Get the current zone name from the interface
		currentZone := zones[i].(map[string]interface{})
		currentZoneName := itemToString(currentZone["name"])
		// If this is the correct zone then get its ID
		if currentZoneName == zoneName {
			zoneID = itemToString(currentZone["id"])
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

// itemToString simply converts an interface item to a string
func itemToString(item interface{}) string {
	return fmt.Sprintf("%v", item)
}

// sendRequest processes the given arguments to send the appropriate request to the Cloudflare API
func (c *Cloudflare) sendRequest(zoneID string, id string, domain string, method string) *http.Response {
	// The base Cloudflare API URL
	url := "https://api.cloudflare.com/client/v4/zones"

	// If the ID and zoneID are not blank then we are retrieving information about a specific record
	if id != "" && zoneID != "" {
		url += "/" + zoneID + "/dns_records/" + id
		// If the ID is blank but the zoneID is not then we are retrieving information about a specific zone
	} else if id == "" && zoneID != "" {
		url += "/" + zoneID + "/dns_records"
	}

	// This processes PUT requests when we need to change a record
	var data []byte
	// Check if it is a PUT request
	if method == "PUT" {
		// Attempt to marshall the given record entry informatino into JSON format
		marshalledData, err := json.Marshal(c.records[domain])
		if err != nil {
			panic(fmt.Errorf("cannot marshall json data for domain %v", domain))
		} else {
			// Store the marshalled JSON data
			data = marshalledData
		}
	} else {
		// If we are not processing a PUT request then set the data portion to null
		data = nil
	}

	// Create a new HTTP client and craft the request with the correct headers
	httpClient := http.Client{}
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
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

	// Return the data from the HTTP request
	return resp
}
