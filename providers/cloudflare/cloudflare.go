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

func (c *Cloudflare) GetIP(entry record.Entry) string {

	c.getInfo(entry)

	response := c.sendRequest(c.records[entry.Domain].ZoneID, c.records[entry.Domain].ID, entry.Domain, "GET")

	var address string

	var recordData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&recordData)
	if itemToString(recordData["success"]) == "true" {
		recordDetails := recordData["result"].(map[string]interface{})
		address = itemToString(recordDetails["content"])
	} else {
		panic(fmt.Errorf("cannot get CloudFlare API response for record %v", entry.Domain))
	}

	return address
}

func (c *Cloudflare) SetIP(address string, entry record.Entry) bool {
	c.getInfo(entry)

	c.records[entry.Domain].Address = address
	response := c.sendRequest(c.records[entry.Domain].ZoneID, c.records[entry.Domain].ID, entry.Domain, "PUT")

	success := false
	var changeData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&changeData)
	if itemToString(changeData["success"]) == "true" {
		success = true
	}

	return success
}

func (c *Cloudflare) getDomainProperty(domain string, property string) string {
	response := c.sendRequest(c.records[domain].ZoneID, "", domain, "GET")

	var value string
	var domainData map[string]interface{}
	json.NewDecoder(response.Body).Decode(&domainData)
	domains := domainData["result"].([]interface{})
	for i := range domains {
		currentDomain := domains[i].(map[string]interface{})
		currentDomainName := itemToString(currentDomain["name"])
		if currentDomainName == domain {
			value = itemToString(currentDomain[property])

		}
	}
	return value
}

func (c *Cloudflare) getInfo(entry record.Entry) {
	_, exists := c.records[entry.Domain]
	if !exists {
		c.records[entry.Domain] = &entry
	}

	if c.records[entry.Domain].ZoneID == "" {
		c.records[entry.Domain].ZoneID = c.getZoneID(entry.Domain)
	}

	if c.records[entry.Domain].ID == "" {
		c.records[entry.Domain].ID = c.getDomainProperty(entry.Domain, "id")
	}

	if c.records[entry.Domain].RecordType == "" {
		c.records[entry.Domain].RecordType = c.getDomainProperty(entry.Domain, "type")
	}

	if c.records[entry.Domain].TTL == 0 {
		ttl, err := strconv.Atoi(c.getDomainProperty(entry.Domain, "ttl"))
		if err == nil {
			c.records[entry.Domain].TTL = ttl
		}
	}

	proxiedS := c.getDomainProperty(entry.Domain, "proxied")
	if proxiedS == "true" {
		c.records[entry.Domain].Proxied = true
	} else {
		c.records[entry.Domain].Proxied = false
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
