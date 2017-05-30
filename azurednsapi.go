/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/* internal implements a stub for the AWS dns API, used primarily for unit testing purposes */
package azuredns

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/dns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// Compile time check for interface conformance
var _ AzureDNSAPI = Api{}

// Interface abstracting the Azure DNS clients from azure-sdk-for-go behind a single interface
// The mock implementation is below
type AzureDNSAPI interface {
	ListZones() (dns.ZoneListResult, error)
	CreateOrUpdateZone(zoneName string, zone dns.Zone, ifMatch string, ifNoneMatch string) (dns.Zone, error)
	DeleteZone(zoneName string, ifMatch string, cancel <-chan struct{}) (result autorest.Response, err error)
	ListResourceRecordSetsByZone(zoneName string) (dns.RecordSetListResult, error)
	CreateOrUpdateRecordSets(zoneName string, relativeRecordSetName string, recordType dns.RecordType, parameters dns.RecordSet, ifMatch string, ifNoneMatch string) (dns.RecordSet, error)
	DeleteRecordSets(zoneName string, relativeRecordSetName string, recordType dns.RecordType, ifMatch string) (result autorest.Response, err error)
}

// AzureDNSAPIStub is a minimal implementation used for unit testing.
// See http://http://docs.aws.amazon.com/sdk-for-go/api/service/dns.html for the full API docs
type Api struct {
	zones      map[string]*dns.Zone
	recordSets map[string][]dns.RecordSet
}

// NewAzureDNSAPIStub returns an initialized AzureDNSAPIStub
func NewAzureDNSAPIStub() *Api {
	api := &Api{
		zones:      make(map[string]*dns.Zone),
		recordSets: make(map[string][]dns.RecordSet),
	}
	api.zones["test.com"] = &dns.Zone{ID: to.StringPtr("zoneID"), Name: to.StringPtr("test.com"), Type: to.StringPtr("ZoneTypes")}
	return api
}

func (a Api) DeleteRecordSets(zoneName string, relativeRecordSetName string, recordType dns.RecordType, ifMatch string) (autorest.Response, error) {
	result := autorest.Response{}

	key := zoneName

	_, ok := a.recordSets[key]
	if ok {
		delete(a.recordSets, key)
	} else {
		// Deleting non-existant item. Some of the tests do that
		return result, nil
	}

	return result, nil
}

func (a Api) CreateOrUpdateRecordSets(zoneName string, relativeRecordSetName string, recordType dns.RecordType, parameters dns.RecordSet, ifMatch string, ifNoneMatch string) (dns.RecordSet, error) {
	var result = parameters
	if _, ok := a.recordSets[zoneName]; ok {
		found := -1

		// we have a zone already ... this might be update
		for i, record := range a.recordSets[zoneName] {
			// check if this record already exists
			if strings.Compare(*record.Type, string(recordType)) == 0 && strings.Compare(*record.Name, *parameters.Name) == 0 {
				found = i
				break
			}
		}
		if found == -1 {
			// zone exists ... record doesn't
			a.recordSets[zoneName] = append(a.recordSets[zoneName], parameters)
		} else {
			if ifNoneMatch == "*" {
				// star parameter says no updates
				return result, fmt.Errorf("parameters don't allow update")
			}

			// zone exists ... record exists
			if *parameters.Etag != "" && a.recordSets[zoneName][found].Etag != parameters.Etag {
				return result, fmt.Errorf("Etag doesn't allow update")
			}
			// update record
			a.recordSets[zoneName][found] = parameters

		}
	} else {
		// new zone
		a.recordSets[zoneName] = make([]dns.RecordSet, 1)
		a.recordSets[zoneName][0] = parameters
		return result, nil
	}

	return result, nil
}

func (a Api) ListResourceRecordSetsByZone(zoneName string) (dns.RecordSetListResult, error) {
	var arr []dns.RecordSet = make([]dns.RecordSet, 0)
	result := dns.RecordSetListResult{}
	result.Value = &arr

	if len(a.recordSets) <= 0 {
		result.Value = &[]dns.RecordSet{}
	} else if _, ok := a.recordSets[zoneName]; !ok {
		result.Value = &[]dns.RecordSet{}
	} else {
		// value is pointer to []RecordSet
		rrset := a.recordSets[zoneName]
		for _, r := range rrset {

			*result.Value = append(*result.Value, dns.RecordSet{Name: r.Name, ID: r.ID, Type: r.Type, RecordSetProperties: r.RecordSetProperties})
		}
	}
	return result, nil
}

func (a Api) ListZones() (dns.ZoneListResult, error) {
	v := make([]dns.Zone, 0)
	result := dns.ZoneListResult{
		Value: &v,
	}
	for _, zone := range a.zones {
		*result.Value = append(*result.Value, *zone)
	}

	return result, nil
}

func (a Api) CreateOrUpdateZone(zoneName string, zone dns.Zone, ifMatch string, ifNoneMatch string) (dns.Zone, error) {
	id := zoneName
	if _, ok := a.zones[id]; ok {
		// zone already exists
		if ifNoneMatch == "*" {
			// update not allowed because of *
			return zone, fmt.Errorf("Error creating hosted DNS zone: %s already exists AND ", id)
		}
		a.zones[id] = &zone
	} else {
		// new zone
		a.zones[id] = &zone
		a.recordSets[id] = make([]dns.RecordSet, 0)
	}

	return zone, nil
}

func (a Api) DeleteZone(zoneName string, ifMatch string, cancel <-chan struct{}) (autorest.Response, error) {
	result := autorest.Response{}

	if len(a.recordSets[zoneName]) > 0 {
		return result, fmt.Errorf("Error deleting hosted DNS zone: %s has resource records", zoneName)
	}

	if z, ok := a.zones[zoneName]; ok {
		if ifMatch == "" || *z.Etag == ifMatch {
			delete(a.zones, zoneName)
		}
	}
	return result, nil
}
