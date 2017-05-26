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
	"strings"
	"github.com/Azure/azure-sdk-for-go/arm/dns"	
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

	result, err := svc.ListResourceRecordSetsByZone( rrsets.zone.Name() )

	if err != nil {
		return nil, err
	}

	var list []dnsprovider.ResourceRecordSet = make([]dnsprovider.ResourceRecordSet, len(*result.Value))

	for i := range *result.Value {
		// value is pointer to []RecordSet 
		var r []dns.RecordSet = *result.Value
		rs := r[i]
		if( &rs != nil ) {
			glog.V(4).Infof("recordset data Name %s Type %s ID %s\n", *rs.Name, *rs.Type, *rs.ID)
			list[i] = &ResourceRecordSet{&(r[i]), &rrsets}
		} else { 
			glog.Fatalf("Recordset was nil\n")
		}
	}

	return list, err
}

func (rrsets ResourceRecordSets) Get(name string) ([]dnsprovider.ResourceRecordSet, error) {
	if( rrsets.zone != nil ) {
		glog.V(5).Infof("GETTING  RecordSets for zone %q, requested %q", rrsets.zone.Name(), name)
	} else {
		glog.V(5).Infof("DANGER GETTING zone is nil\n")
		
	}
	rrsetList, err := rrsets.List()
	arr := make([]dnsprovider.ResourceRecordSet, 0) 

	if err != nil {
		return nil, err
	}
	for _, rrset := range rrsetList {
		glog.V(5).Infof("azuredns: ResourceRecrdSets Get looking for %q found %q\n", name, rrset.Name())
		if rrset.Name() == name {
			arr = append( arr, rrsets.New( rrset.Name(), rrset.Rrdatas(), rrset.Ttl(), rrset.Type() ) )
		}
	}

	glog.V(5).Infof("azuredns: ResourceRecrdSets Get for %q returned %i records\n", name, len(arr))
	if len(arr) <= 0 {
		return nil, nil
	}
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

	relativeName := strings.TrimSuffix( name, *r.zone.impl.Name )
	relativeName = strings.TrimSuffix( relativeName, ".")
	rs := &dns.RecordSet{
		Name: &relativeName,
		Type: &rrstypeStr,
	}
	
	rrs := ResourceRecordSet{
		rs,
		&r,
	}
	return rrs.setRecordSetProperties(ttl, rrdatas)
}

// Zone returns the parent zone
func (rrset ResourceRecordSets) Zone() dnsprovider.Zone {
	return rrset.zone
}
