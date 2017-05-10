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
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/glog"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

// Compile time check for interface adherence
var _ dnsprovider.Zones = Zones{}

type Zones struct {
	interface_ *Interface
}

func (zones Zones) List() ([]dnsprovider.Zone, error) {
	var zoneList []dnsprovider.Zone
	svc := *zones.interface_.service
	// request all 100 zones. 100 is the current limit per subscription
	azZoneList, err := svc.GetZonesClient().List(to.Int32Ptr(100))
	if err == nil {
		fmt.Printf("got %i zones\n", len(*azZoneList.Value))
		for _, zone := range *azZoneList.Value {
			fmt.Printf("Adding zone %s to list\n", *zone.Name)

			// TODO: Get Resource Records for this zone!

			zoneList = append(zoneList, &Zone{&zone, &zones})
		}
	} else {
		fmt.Printf("Error listing: %s\n", err.Error)
	}

	return zoneList, nil
}

func (zones Zones) Add(zone dnsprovider.Zone) (dnsprovider.Zone, error) {
	zoneName := zone.Name()
	svc := *zones.interface_.service
	zoneParam := &dns.Zone{
		Location: to.StringPtr("global"),
		Name:     to.StringPtr(zoneName),
	}

	fmt.Printf("Creating Zone: %s, in resource group: %s\n", zoneName, svc.GetResourceGroupName())
	_, err := svc.GetZonesClient().CreateOrUpdate(
		svc.GetResourceGroupName(),
		zoneName, *zoneParam, "", "")

	if err != nil {
		glog.Errorf("Error creating Azure DNS zone: %s: %s", zoneName, err.Error)
		return nil, err
	}

	return &Zone{
		impl:  zoneParam,
		zones: &zones}, nil
}

func (zones Zones) Remove(zone dnsprovider.Zone) error {
	svc := *zones.interface_.service

	fmt.Printf("Removing Azure DNS zone ID: %s, Name: %s rg: %s\n", zone.ID(), zone.Name(), svc.GetResourceGroupName())
	_, err := svc.GetZonesClient().Delete(svc.GetResourceGroupName(),
		zone.Name(), "", nil)
	// config := NewAzureDNSConfig("ffa90503-6fe8-4ab5-9bf1-059f81a6d8bb",
	// 	"delete-dns",
	// 	"66841164-1e0e-4ffc-a0d2-ce36f95ec41d",
	// 	"eebec4ca-c175-45dc-b763-943607ce4838",
	// 	"9f4ba4e0-be35-4821-9aa6-4caadfaba4fa")
	// c := map[string]string{
	// 	"AZURE_CLIENT_ID":       config.clientId,
	// 	"AZURE_CLIENT_SECRET":   config.secret,
	// 	"AZURE_SUBSCRIPTION_ID": config.subscriptionId,
	// 	"AZURE_TENANT_ID":       config.tenantId}
	// zc := dns.NewZonesClient(config.subscriptionId)
	// spt, err := helpers.NewServicePrincipalTokenFromCredentials(c, azure.PublicCloud.ResourceManagerEndpoint)
	// if err != nil {
	// 	glog.Fatalf("Error authenticating to Azure DNS: %v", err)
	// 	return nil
	// }

	// zc.Authorizer = autorest.NewBearerAuthorizer(spt)

	// _, newerr := zc.Delete(svc.GetResourceGroupName(),
	// 	zone.Name(), "", nil)
	if err != nil {
		//fmt.Printf("type: %s\n", newerr.(Type))
		// TODO: Fix go azure sdk version
		fmt.Printf("Error Deleting Azure DNS Zone %s %v\n", zone.Name(), err.Error())
		// e := <-newerr
		// fmt.Printf("Error Deleting Azure DNS Zone %s %v\n", zone.Name(), e.Error())
		// fmt.Fprintf()
		// TODO --- massively
		//newErr
		return err
	}
	return nil
}
func (zones Zones) New(name string) (dnsprovider.Zone, error) {
	zone := dns.Zone{ID: &name, Name: &name}

	fmt.Printf("New Zone: %s, ID %s was name %s\n", *zone.Name, *zone.ID, name)
	return &Zone{&zone, &zones}, nil
}
