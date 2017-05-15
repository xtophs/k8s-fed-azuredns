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
	"bytes"
	"fmt"
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
		recs := make([]dns.ARecord, len(rrDatas))

		for i = range rrDatas {
			recs[i] = dns.ARecord{
				Ipv4Address: to.StringPtr(rrDatas[i]),
			}
		}
		props.ARecords = &recs

	case "AAAA":
		recs := make([]dns.AaaaRecord, len(rrDatas))
		for i = range rrDatas {
			recs[i] = dns.AaaaRecord{
				Ipv6Address: to.StringPtr(rrDatas[i]),
			}
		}
		props.AaaaRecords = &recs

	case "CNAME":
		for i = range rrDatas {
			props.CnameRecord = &dns.CnameRecord{
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
		rrDatas = make([]string, len(*props.AaaaRecords))

		for i := range *props.AaaaRecords {
			rec := *props.AaaaRecords
			rrDatas[i] = *rec[i].Ipv6Address
		}

	case "CNAME":
		rrDatas = make([]string, 1)
		rrDatas[0] = *props.CnameRecord.Cname
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
		if glog.V(8) {
			var sb bytes.Buffer
			sb.WriteString(fmt.Sprintf("\tRecordSet: %s Type: %s Zone Name %s\n", *rrset.Name, *recType, *zoneName))
			glog.V(8).Infof("Azure DNS Removal:\n%s", sb.String())
		}

		_, err := svc.RecordSetsDelete( *zoneName, *rrset.Name, dns.RecordType(*recType), "")

		if err != nil {
			glog.V(8).Infof("Could not delete DNS %s", *rrset.Name)
			return err
		}


	}

	for _, upsert := range c.upserts {
		var rrset = fromProviderRrset(upsert)
		recType := rrset.Type
		if glog.V(8) {
			var sb bytes.Buffer
			sb.WriteString(fmt.Sprintf("\tRecordSet: %s Type: %s Zone Name %s\n", *rrset.Name, *recType, *zoneName))
			glog.V(8).Infof("Azure DNS Upsert:\n%s", sb.String())
		}

		_, err := svc.RecordSetCreateOrUpdate(	*zoneName, *rrset.Name, dns.RecordType(*recType), *rrset, "", "*")

		if err != nil {
			glog.V(0).Infof("Could not upsert DNS %s", upsert.Name)
			return err
		}
	}

	for _, addition := range c.additions {
		var rrset = fromProviderRrset(addition)
		recType := rrset.Type

		if glog.V(8) {
			var sb bytes.Buffer
			sb.WriteString(fmt.Sprintf("\tRecordSet: %s Type: %s Zone Name %s\n", *rrset.Name, *recType, *zoneName))
			glog.V(8).Infof("Azure DNS Addition:\n%s", sb.String())
		}

		_, err := svc.RecordSetCreateOrUpdate(*zoneName, *rrset.Name, dns.RecordType(*recType), *rrset, "", "*")
		if err != nil {
			glog.V(0).Infof("Coul not add DNS %s", addition.Name)
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
