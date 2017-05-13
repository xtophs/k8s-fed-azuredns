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
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/rrstype"
	"github.com/golang/glog"
)

// Compile time check for interface adherence
var _ dnsprovider.ResourceRecordSets = ResourceRecordSets{}

type ResourceRecordSets struct {
	zone *Zone
}

func (rrsets ResourceRecordSets) List() ([]dnsprovider.ResourceRecordSet, error) {

	svc := *rrsets.zone.zones.interface_.service
	glog.V(5).Infof("LISTING RecordSets for zone %s in rg %s\n", rrsets.zone.Name(), svc.GetResourceGroupName())

	result, err := svc.GetRecordSetsClient().ListByDNSZone(
		svc.GetResourceGroupName(),
		rrsets.zone.Name(),
		to.Int32Ptr(100))

	if err != nil {
		return nil, err
	}

	// TODO: paging
	var list []dnsprovider.ResourceRecordSet = make([]dnsprovider.ResourceRecordSet, len(*result.Value))

	for i := range *result.Value {
		var r []dns.RecordSet = *result.Value
		glog.V(5).Infof("recordset data: %s, %i\n", *r[i].Name, *r[i].TTL)
		list[i] = &ResourceRecordSet{&(r[i]), &rrsets}
	}

	return list, err
}

func (rrsets ResourceRecordSets) Get(name string) ([]dnsprovider.ResourceRecordSet, error) {
	glog.V(5).Infof("GETTING  RecordSets for zone %s\n", rrsets.zone.Name())
	var newRrset dnsprovider.ResourceRecordSet
	rrsetList, err := rrsets.List()
	if err != nil {
		return nil, err
	}
	for _, rrset := range rrsetList {
		if rrset.Name() == name {
			newRrset = rrset
			break
		}
	}
	arr := make([]dnsprovider.ResourceRecordSet, 1) 
	arr[0] = newRrset
	return arr, nil
}

func (r ResourceRecordSets) StartChangeset() dnsprovider.ResourceRecordChangeset {
	return &ResourceRecordChangeset{
		zone:   r.zone,
		rrsets: &r,
	}
}

func (r ResourceRecordSets) New(name string, rrdatas []string, ttl int64, rrstype rrstype.RrsType) dnsprovider.ResourceRecordSet {
	rrstypeStr := string(rrstype)
	rrs := &dns.RecordSet{
		Name: &name,
		Type: &rrstypeStr,
	}

	rrs.RecordSetProperties = rrDatasToRecordSetProperties(rrstypeStr, rrdatas)

	return ResourceRecordSet{
		rrs,
		&r,
	}
}

// Zone returns the parent zone
func (rrset ResourceRecordSets) Zone() dnsprovider.Zone {
	return rrset.zone
}
