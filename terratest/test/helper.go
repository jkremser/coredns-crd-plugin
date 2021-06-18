/*
Copyright 2021 ABSA Group Limited

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Generated by GoLic, for more details see: https://github.com/AbsaOSS/golic
*/
package test

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

const (
	DefaultTimeout time.Duration = 5 * time.Second
)

type clientIP struct {
	ip string
	Opts *dns.OPT
}

func NewClientIP(ip string) (cip *clientIP) {
	cip = new(clientIP)
	cip.ip = ip
	cip.Opts = &dns.OPT{}
	if !cip.IsEmpty() {
		subnet := &dns.EDNS0_SUBNET{
			Code:          dns.EDNS0SUBNET,
			Address:       net.ParseIP(ip),
			Family:        1, // IP4
			SourceNetmask: net.IPv4len * 8,
		}
		cip.Opts = &dns.OPT{
			Hdr: dns.RR_Header{
				Name:   ".",
				Rrtype: dns.TypeOPT,
			},
			Option: []dns.EDNS0{subnet},
		}
		return cip
	}
	return
}

func (cip *clientIP) IsEmpty() bool {
	return cip.ip == ""
}


func queryDNS(dnsServer string, dnsPort int, dnsName string, dnsType uint16, clientIP *clientIP) (*dns.Msg, error) {
	dnsName = fmt.Sprintf("%s.", dnsName)
	m := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			RecursionDesired: true,
		},
		Question: make([]dns.Question, 1),
	}
	c := &dns.Client{
		ReadTimeout: DefaultTimeout,
	}
	if !clientIP.IsEmpty() {
		m.Extra = append(m.Extra, clientIP.Opts)
	}
	c.Net = "udp4"
	c.Dialer = &net.Dialer{}

	m.SetQuestion(dnsName, dnsType)
	r, _, err := c.Exchange(m, fmt.Sprintf("%s:%d", dnsServer, dnsPort))

	return r, err
}

func DigMsg(t *testing.T, dnsServer string, dnsPort int, dnsName string, dnsType uint16) (*dns.Msg, error) {
	return queryDNS(dnsServer, dnsPort, dnsName, dnsType, NewClientIP(""))
}

func DigIPs(t *testing.T, dnsServer string, dnsPort int, dnsName string, dnsType uint16, clientIP string) ([]string, error) {
	var result []string
	r, err := queryDNS(dnsServer, dnsPort, dnsName, dnsType, NewClientIP(clientIP))

	if err != nil {
		return nil, err
	}

	for _, record := range r.Answer {
		if e, ok := record.(*dns.A); ok {
			if e.A == nil {
				return nil, errors.New("malformed message packet")
			}
			result = append(result, e.A.String())
		}
	}
	return result, nil
}
