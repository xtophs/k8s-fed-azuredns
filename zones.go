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
	svc := *zones.interface_.service

	azZoneList, err := svc.ListZones()
	var zoneList []dnsprovider.Zone

	if err == nil {
		glog.V(5).Infof("got %i zones\n", len(*azZoneList.Value))
		for _, zone := range *azZoneList.Value {
			glog.V(5).Infof("got %v\n", zone)

			zoneList = append( zoneList, &Zone{&zone, &zones}) 
		}
	} else {
		glog.V(5).Infof("Error listing: %s\n", err.Error())
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

	_, err := svc.CreateOrUpdateZone(zoneName, *zoneParam, "", "")

	if err != nil {
		glog.Errorf("Error creating Azure DNS zone: %s: %s", zoneName, err.Error())
		return nil, err
	}

	return &Zone{
		impl:  zoneParam,
		zones: &zones}, nil
}

func (zones Zones) Remove(zone dnsprovider.Zone) error {
	svc := *zones.interface_.service
	_, err := svc.DeleteZone(zone.Name(), "", nil)

	if err != nil {
		// TODO: Fix go azure sdk version
		// Is now async
		glog.V(0).Infof("Error Deleting Azure DNS Zone %s %v\n", zone.Name(), err.Error())
		return err
	}
	return nil
}
func (zones Zones) New(name string) (dnsprovider.Zone, error) {
	zone := dns.Zone{ID: &name, Name: &name}
	return &Zone{&zone, &zones}, nil
}
