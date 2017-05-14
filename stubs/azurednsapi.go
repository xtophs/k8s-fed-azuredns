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
package stubs

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/dns"
	//	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
)

// Compile time check for interface conformance
var _ AzureDNSAPI = &AzureDNSAPIStub{}

/* dnsAPI is the subset of the AWS dns API that we actually use.  Add methods as required. Signatures must match exactly. */
type AzureDNSAPI interface {
	// 	ListResourceRecordSetsPages(input *dns.ListResourceRecordSetsInput, fn func(p *dns.ListResourceRecordSetsOutput, lastPage bool) (shouldContinue bool)) error
	// 	ChangeResourceRecordSets(*dns.ChangeResourceRecordSetsInput) (*dns.ChangeResourceRecordSetsOutput, error)
	// 	ListHostedZonesPages(input *dns.ListHostedZonesInput, fn func(p *dns.ListHostedZonesOutput, lastPage bool) (shouldContinue bool)) error
	// 	CreateHostedZone(*dns.CreateHostedZoneInput) (*dns.CreateHostedZoneOutput, error)
	// 	DeleteHostedZone(*dns.DeleteHostedZoneInput) (*dns.DeleteHostedZoneOutput, error)
}

// AzureDNSAPIStub is a minimal implementation of dnsAPI, used primarily for unit testing.
// See http://http://docs.aws.amazon.com/sdk-for-go/api/service/dns.html for descriptions
// of all of its methods.
type AzureDNSAPIStub struct {
	zc dns.ZonesClient
	rc dns.RecordSetsClient
	rg string
}

// NewAzureDNSAPIStub returns an initialized AzureDNSAPIStub
func NewAzureDNSAPIStub() *AzureDNSAPIStub {
	//  TODO
	return &AzureDNSAPIStub{
	// zones:      make(map[string]*dns.Zone),
	// recordSets: make(map[string]map[string][]*dns.RecordSet),
	}
}

func (api *AzureDNSAPIStub) RecordSetCreateOrUpdate(resourceGroupName string, zoneName string, parameters dns.Zone, ifMatch string, ifNoneMatch string) (result dns.Zone, err error) {
	// TODO
	fmt.Printf("Testing ZonesClient CreateOrUpdate")
	return dns.Zone{}, nil
}

func (a *AzureDNSAPIStub) ListZones() []dns.Zones {
	fake := FakeZonesClient{}
	c := dns.ZonesClient(fake)
	return &c // *dns.ZonesClient(&FakeZonesClient{})
}

func (a *AzureDNSAPIStub) GetRecordSetsClient() *dns.RecordSetsClient {
	fake := FakeRecordSetsClient{}

	var c dns.RecordSetsClient
	c = dns.RecordSetsClient(fake)

	return &c
}

func (a AzureDNSAPIStub) GetResourceGroupName() string {
	return "fakeResourceGroup"
}

// func (r *AzureDNSAPIStub) ListResourceRecordSetsPages(input *dns.ListResourceRecordSetsInput, fn func(p *dns.ListResourceRecordSetsOutput, lastPage bool) (shouldContinue bool)) error {
// 	output := dns.ListResourceRecordSetsOutput{} // TODO: Support optional input args.
// 	if len(r.recordSets) <= 0 {
// 		output.ResourceRecordSets = []*dns.ResourceRecordSet{}
// 	} else if _, ok := r.recordSets[*input.HostedZoneId]; !ok {
// 		output.ResourceRecordSets = []*dns.ResourceRecordSet{}
// 	} else {
// 		for _, rrsets := range r.recordSets[*input.HostedZoneId] {
// 			for _, rrset := range rrsets {
// 				output.ResourceRecordSets = append(output.ResourceRecordSets, rrset)
// 			}
// 		}
// 	}
// 	lastPage := true
// 	fn(&output, lastPage)
// 	return nil
// }

// func (r *AzureDNSAPIStub) ChangeResourceRecordSets(input *dns.ChangeResourceRecordSetsInput) (*dns.ChangeResourceRecordSetsOutput, error) {
// 	output := &dns.ChangeResourceRecordSetsOutput{}
// 	recordSets, ok := r.recordSets[*input.HostedZoneId]
// 	if !ok {
// 		recordSets = make(map[string][]*dns.ResourceRecordSet)
// 	}

// 	for _, change := range input.ChangeBatch.Changes {
// 		key := *change.ResourceRecordSet.Name + "::" + *change.ResourceRecordSet.Type
// 		switch *change.Action {
// 		case dns.ChangeActionCreate:
// 			if _, found := recordSets[key]; found {
// 				return nil, fmt.Errorf("Attempt to create duplicate rrset %s", key) // TODO: Return AWS errors with codes etc
// 			}
// 			recordSets[key] = append(recordSets[key], change.ResourceRecordSet)
// 		case dns.ChangeActionDelete:
// 			if _, found := recordSets[key]; !found {
// 				return nil, fmt.Errorf("Attempt to delete non-existent rrset %s", key) // TODO: Check other fields too
// 			}
// 			delete(recordSets, key)
// 		case dns.ChangeActionUpsert:
// 			// TODO - not used yet
// 		}
// 	}
// 	r.recordSets[*input.HostedZoneId] = recordSets
// 	return output, nil // TODO: We should ideally return status etc, but we don't' use that yet.
// }

// func (r *AzureDNSAPIStub) ListHostedZonesPages(input *dns.ListHostedZonesInput, fn func(p *dns.ListHostedZonesOutput, lastPage bool) (shouldContinue bool)) error {
// 	output := &dns.ListHostedZonesOutput{}
// 	for _, zone := range r.zones {
// 		output.HostedZones = append(output.HostedZones, zone)
// 	}
// 	lastPage := true
// 	fn(output, lastPage)
// 	return nil
// }

// func (r *AzureDNSAPIStub) CreateHostedZone(input *dns.CreateHostedZoneInput) (*dns.CreateHostedZoneOutput, error) {
// 	name := aws.StringValue(input.Name)
// 	id := "/hostedzone/" + name
// 	if _, ok := r.zones[id]; ok {
// 		return nil, fmt.Errorf("Error creating hosted DNS zone: %s already exists", id)
// 	}
// 	r.zones[id] = &dns.HostedZone{
// 		Id:   aws.String(id),
// 		Name: aws.String(name),
// 	}
// 	return &dns.CreateHostedZoneOutput{HostedZone: r.zones[id]}, nil
// }

// func (r *AzureDNSAPIStub) DeleteHostedZone(input *dns.DeleteHostedZoneInput) (*dns.DeleteHostedZoneOutput, error) {
// 	if _, ok := r.zones[*input.Id]; !ok {
// 		return nil, fmt.Errorf("Error deleting hosted DNS zone: %s does not exist", *input.Id)
// 	}
// 	if len(r.recordSets[*input.Id]) > 0 {
// 		return nil, fmt.Errorf("Error deleting hosted DNS zone: %s has resource records", *input.Id)
// 	}
// 	delete(r.zones, *input.Id)
// 	return &dns.DeleteHostedZoneOutput{}, nil
// }
