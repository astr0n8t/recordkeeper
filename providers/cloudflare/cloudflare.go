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

type Cloudflare struct {
	username  string
	authToken string
	records   map[string]*record.Entry
}

func New(user string, auth string) *Cloudflare {
	return &Cloudflare{user, auth, make(map[string]*record.Entry)}
}

func (c *Cloudflare) GetIP(domain string) string {

	c.getInfo(domain)

	response := c.sendRequest(c.records[domain].ZoneID, c.records[domain].ID, domain, "GET")

	var address string

	var recordData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&recordData)
	if itemToString(recordData["success"]) == "true" {
		recordDetails := recordData["result"].(map[string]interface{})
		address = itemToString(recordDetails["content"])
	} else {
		panic(fmt.Errorf("cannot get CloudFlare API response for record %v", domain))
	}

	return address
}

func (c *Cloudflare) SetIP(domain string, address string) bool {
	c.getInfo(domain)

	c.records[domain].Address = address
	response := c.sendRequest(c.records[domain].ZoneID, c.records[domain].ID, domain, "PUT")

	success := false
	var changeData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&changeData)
	if itemToString(changeData["success"]) == "true" {
		success = true
	}

	return success
}

func (c *Cloudflare) getDomainInfo(domain string) {
	response := c.sendRequest(c.records[domain].ZoneID, "", domain, "GET")

	var domainData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&domainData)
	domains := domainData["result"].([]interface{})
	for i := range domains {
		currentDomain := domains[i].(map[string]interface{})
		currentDomainName := itemToString(currentDomain["name"])
		if currentDomainName == domain {
			c.records[domain].ID = itemToString(currentDomain["id"])
			c.records[domain].RecordType = itemToString(currentDomain["type"])
			ttl, err := strconv.Atoi(itemToString(currentDomain["ttl"]))
			if err == nil {
				c.records[domain].TTL = ttl
			}
			proxiedS := itemToString(currentDomain["proxied"])
			if proxiedS == "true" {
				c.records[domain].Proxied = true
			} else {
				c.records[domain].Proxied = false
			}
		}
	}
}

func (c *Cloudflare) getInfo(domain string) {
	domainRecord, exists := c.records[domain]
	if !exists {
		domainRecord = &record.Entry{domain, "", "", c.getZoneID(domain), "", false, 0}
		c.records[domain] = domainRecord
		c.getDomainInfo(domain)
	}
}

func (c *Cloudflare) getZoneID(domain string) string {
	var zoneID string
	zoneName := findZoneName(domain)
	response := c.sendRequest("", "", domain, "GET")

	var zoneData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&zoneData)
	zones := zoneData["result"].([]interface{})
	for i := range zones {
		currentZone := zones[i].(map[string]interface{})
		currentZoneName := itemToString(currentZone["name"])
		if currentZoneName == zoneName {
			zoneID = itemToString(currentZone["id"])
		}
	}

	return zoneID
}

func findZoneName(domain string) string {
	zoneNameSlice := strings.Split(domain, ".")

	var zoneName string
	if len(zoneNameSlice) > 1 {
		zoneName = zoneNameSlice[len(zoneNameSlice)-2] + "." + zoneNameSlice[len(zoneNameSlice)-1]
	} else {
		zoneName = domain
	}

	return zoneName
}

func itemToString(item interface{}) string {
	return fmt.Sprintf("%v", item)
}

func (c *Cloudflare) sendRequest(zoneID string, id string, domain string, method string) *http.Response {
	url := "https://api.cloudflare.com/client/v4/zones"

	if id != "" && zoneID != "" {
		url += "/" + zoneID + "/dns_records/" + id
	} else if id == "" && zoneID != "" {
		url += "/" + zoneID + "/dns_records"
	}

	var data []byte
	if method == "PUT" {
		marshalledData, err := json.Marshal(c.records[domain])
		if err != nil {
			panic(fmt.Errorf("cannot marshall json data for domain %v", domain))
		} else {
			data = marshalledData
		}
	} else {
		data = nil
	}

	httpClient := http.Client{}
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	request.Header.Add("X-Auth-Email", c.username)
	request.Header.Add("X-Auth-Key", c.authToken)
	request.Header.Add("X-Content-Type", "application/json")

	if err != nil {
		panic(fmt.Errorf("cannot form CloudFlare API request"))
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		panic(fmt.Errorf("cannot connect to CloudFlare API"))
	}

	return resp
}
