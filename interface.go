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

package azuredns

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/dns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"github.com/Azure/go-autorest/autorest/to"
)

// Compile time check for interface adherence
var _ dnsprovider.Interface = Interface{}

var _ AzureDNSAPI = &Clients{}

type Interface struct {
	service *AzureDNSAPI
}

type Clients struct {
	rc      dns.RecordSetsClient
	zc      dns.ZonesClient
	conf    Config
}

func( c *Clients) DeleteRecordSets(zoneName string, relativeRecordSetName string, recordType dns.RecordType, ifMatch string) (result autorest.Response, err error){
	glog.V(4).Infof("azuredns: Deleting RecordSet %q type %q for zone %s in rg %q\n", relativeRecordSetName, string(recordType), zoneName, c.conf.Global.ResourceGroup)

	return c.rc.Delete(c.conf.Global.ResourceGroup, zoneName, relativeRecordSetName, recordType, ifMatch) 
}

func( c *Clients) CreateOrUpdateRecordSets(zoneName string, relativeRecordSetName string, recordType dns.RecordType, parameters dns.RecordSet, ifMatch string, ifNoneMatch string) (dns.RecordSet, error) {
	glog.V(4).Infof("azuredns: CreateOrUpdate RecordSets %q type %q for zone %q in rg %q\n", relativeRecordSetName, string(recordType), zoneName, c.conf.Global.ResourceGroup)

	return c.rc.CreateOrUpdate(c.conf.Global.ResourceGroup, 
		zoneName, relativeRecordSetName , recordType, parameters, ifMatch, ifNoneMatch) 
}

func( c *Clients) ListResourceRecordSetsByZone(zoneName string )(dns.RecordSetListResult, error)  {
	glog.V(5).Infof("azuredns: Listing RecordSets for zone %s in rg %s\n", zoneName, c.conf.Global.ResourceGroup)

	//var records []dns.RecordSet = make([]dns.RecordSet, 0)

	// TODO: paging
	result, err := c.rc.ListByDNSZone(	c.conf.Global.ResourceGroup,
		zoneName,
		to.Int32Ptr(1000) )

	// for _, r := range *result.Value {
	// 	records = append(records, r)
	// }


	// if result.NextLink != nil {
	// 	if *result.NextLink != "" {
	// 		result, err := c.rc.ListByDNSZone(	c.conf.Global.ResourceGroup,
	// 			zoneName,
	// 			to.Int32Ptr(10))

	// 		for _, r := range *result.Value {
	// 			records = append(records, r)
	// 		}
	// 	}
	// }

	//*result.Value = records 

	return result, err
}

func( c *Clients ) ListZones() ( dns.ZoneListResult, error) {
	glog.V(5).Infof("azuredns: Requesting DNS zones")
	// request all 100 zones. 100 is the current limit per subscription
	return c.zc.List( to.Int32Ptr(100))
}

func( c *Clients ) CreateOrUpdateZone( zoneName string, zone dns.Zone, ifMatch string, ifNoneMatch string ) (  dns.Zone, error) {
	glog.V(4).Infof("azuredns: Creating Zone: %s, in resource group: %s\n", zoneName, c.conf.Global.ResourceGroup)
	return c.zc.CreateOrUpdate(c.conf.Global.ResourceGroup, zoneName, zone, ifMatch, ifNoneMatch )
}

func( c *Clients ) DeleteZone( zoneName string, ifMatch string, cancel <-chan struct{}) (result autorest.Response, err error){
	glog.V(4).Infof("azuredns: Removing Azure DNS zone Name: %s rg: %s\n", zoneName, c.conf.Global.ResourceGroup)
	return c.zc.Delete( c.conf.Global.ResourceGroup, zoneName, ifMatch,  cancel)
}

// New builds an Interface, with a specified azurednsAPI implementation.
// This is useful for testing purposes, but also if we want an instance with with custom AWS options.
func New(service *AzureDNSAPI) *Interface {
	return &Interface{service}
}

func NewAPI(config Config) *Interface {
	var clients *Clients
	clients = &Clients{}			

	glog.V(4).Infof("azuredns: Created Azure DNS clients for subscription: %s", config.Global.SubscriptionID)

	clients.conf = config

	clients.zc = dns.NewZonesClient(config.Global.SubscriptionID)
	spt, err := NewServicePrincipalTokenFromCredentials(config, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		glog.Fatalf("azuredns: Error authenticating to Azure DNS: %v", err)
		return nil
	}

	clients.zc.Authorizer = spt

	clients.rc = dns.NewRecordSetsClient(config.Global.SubscriptionID)
	spt, err = NewServicePrincipalTokenFromCredentials(config, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		glog.Fatalf("azuredns: Error authenticating to Azure DNS: %v", err)
		return nil
	}

	clients.rc.Authorizer = spt
	var api AzureDNSAPI = clients
	return &Interface{&api}
}

func (i Interface) Zones() (zones dnsprovider.Zones, supported bool) {
	return Zones{&i}, true
}

func checkEnvVar(envVars *map[string]string) error {
	var missingVars []string
	for varName, value := range *envVars {
		if value == "" {
			missingVars = append(missingVars, varName)
		}
	}
	if len(missingVars) > 0 {
		return fmt.Errorf("azuredns: Missing environment variables %v", missingVars)
	}
	return nil
}
