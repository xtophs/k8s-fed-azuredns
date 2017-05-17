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
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/glog"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

// Compile time check for interface adherence
var _ dnsprovider.ResourceRecordChangeset = &ResourceRecordChangeset{}

type ResourceRecordChangeset struct {
	zone   *Zone
	rrsets *ResourceRecordSets

	additions []dnsprovider.ResourceRecordSet
	removals  []dnsprovider.ResourceRecordSet
	upserts   []dnsprovider.ResourceRecordSet
}

func (c *ResourceRecordChangeset) Add(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.additions = append(c.additions, rrset)
	return c
}

func (c *ResourceRecordChangeset) Remove(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.removals = append(c.removals, rrset)
	return c
}

func (c *ResourceRecordChangeset) Upsert(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.upserts = append(c.upserts, rrset)
	return c
}

// TODO
func rrDatasToRecordSetProperties(rrsType string, rrDatas []string) *dns.RecordSetProperties {
	props := dns.RecordSetProperties{}
	var i int
	// kubernetes 1.6.2 only handles A, AAAA and CNAME
	switch rrsType {
	case "A":
		recs := make([]dns.ARecord, 1)
			recs[0] = dns.ARecord{
				Ipv4Address: to.StringPtr(rrDatas[0]),
			}

		// for i = range rrDatas {
		// 	recs[i] = dns.ARecord{
		// 		Ipv4Address: to.StringPtr(rrDatas[i]),
		// 	}
		// }
		props.ARecords = &recs

	case "AAAA":
		recs := make([]dns.AaaaRecord, 1)
		recs[0] = dns.AaaaRecord{
			Ipv6Address: to.StringPtr(rrDatas[0]),
		}

		// for i = range rrDatas {
		// 	recs[i] = dns.AaaaRecord{
		// 		Ipv6Address: to.StringPtr(rrDatas[i]),
		// 	}
		// }
		props.AAAARecords = &recs

	case "CNAME":
		for i = range rrDatas {
			props.CNAMERecord = &dns.CnameRecord{
				Cname: to.StringPtr(rrDatas[i]),
			}
		}
	}

	return &props
}

func recordSetPropertiesToRrDatas(rset *dns.RecordSet) []string {

	props := rset.RecordSetProperties
	var rrDatas []string

	switch strings.TrimPrefix(*rset.Type, "Microsoft.Network/dnszones/") {
	case "A":
		rrDatas = make([]string, len(*props.ARecords))

		for i := range *props.ARecords {
			rec := *props.ARecords
			rrDatas[i] = *rec[i].Ipv4Address
		}

	case "AAAA":
		rrDatas = make([]string, len(*props.AAAARecords))

		for i := range *props.AAAARecords {
			rec := *props.AAAARecords
			rrDatas[i] = *rec[i].Ipv6Address
		}

	case "CNAME":
		rrDatas = make([]string, 1)
		rrDatas[0] = *props.CNAMERecord.Cname
	}

	return rrDatas
}

func fromProviderRrset(rrset dnsprovider.ResourceRecordSet) *dns.RecordSet {
	recType := string(rrset.Type())

	changeRecord := &dns.RecordSet{
		Name: to.StringPtr(rrset.Name()),
		Type: to.StringPtr(recType),
	}
	changeRecord.RecordSetProperties = rrDatasToRecordSetProperties(
		recType, rrset.Rrdatas())

	changeRecord.RecordSetProperties.TTL = to.Int64Ptr(rrset.Ttl())
	return changeRecord
}

func (c *ResourceRecordChangeset) Apply() error {

	zoneName := c.zone.impl.Name
	// TODO
	// can I combine requests into a batch?
	// since it looks like the autorest API is request/response we can
	// start with calling the REST APIs one-by-one

	svc := *c.rrsets.zone.zones.interface_.service

	for _, removal := range c.removals {
		var rrset = fromProviderRrset(removal)
		recType := rrset.Type
		// TODO Refactor
		glog.V(1).Infof("azuredns: Delete:\tRecordSet: %s Type: %s Zone Name: %s TTL: %i \n", *rrset.Name, *recType, *zoneName, *rrset.RecordSetProperties.TTL)

		_, err := svc.DeleteRecordSets( *zoneName, *rrset.Name, dns.RecordType(*recType), "")

		if err != nil {
			glog.V(1).Infof("azuredns: Could not delete DNS %s", *rrset.Name)
			return err
		}


	}

	for _, upsert := range c.upserts {
		var rrset = fromProviderRrset(upsert)
		recType := rrset.Type
		glog.V(1).Infof("azuredns: Upsert:\tRecordSet: %s Type: %s Zone Name: %s TTL: %i \n", *rrset.Name, *recType, *zoneName, *rrset.RecordSetProperties.TTL)

		_, err := svc.CreateOrUpdateRecordSets(	*zoneName, *rrset.Name, dns.RecordType(*recType), *rrset, "", "*")

		if err != nil {
			glog.V(0).Infof("azuredns: Could not upsert DNS %s", upsert.Name)
			return err
		}
	}

	for _, addition := range c.additions {
		var rrset = fromProviderRrset(addition)
		recType := rrset.Type

		glog.V(0).Infof("azuredns:  Addition:\tRecordSet: %s Type: %s Zone Name: %s TTL: %i \n", *rrset.Name, *recType, *zoneName, *rrset.RecordSetProperties.TTL)

		props := rrset.RecordSetProperties
		glog.V(0).Infof("Type %s\n",rrset.Type)
		switch strings.TrimPrefix(*recType, "Microsoft.Network/dnszones/") {	
		case "A":
			for i := range *props.ARecords {
				rec := *props.ARecords
				glog.V(0).Infof("A Rec Ipv4: %s\n", *rec[i].Ipv4Address)
			}

		case "AAAA":
			for i := range *props.AAAARecords {
				rec := *props.AAAARecords
				glog.V(0).Infof("AAAA Rec Ipv6: %s\n", *rec[i].Ipv6Address)
			}

		case "CNAME":
			glog.V(0).Infof("CNAME: %s\n", *props.CNAMERecord.Cname)
		}

		
		_, err := svc.CreateOrUpdateRecordSets(*zoneName, *rrset.Name, dns.RecordType(*recType), *rrset, "", "*")
		if err != nil {
			glog.V(0).Infof("azuredns: Could not add DNS %s: %s", addition.Name(), err.Error())
			return err
		}
	}

	return nil
}

func (c *ResourceRecordChangeset) IsEmpty() bool {
	return len(c.removals) == 0 && len(c.additions) == 0 && len(c.upserts) == 0
}

// ResourceRecordSets returns the parent ResourceRecordSets
func (c *ResourceRecordChangeset) ResourceRecordSets() dnsprovider.ResourceRecordSets {
	return c.rrsets
}
