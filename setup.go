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
package gateway

import (
	"context"

	"fmt"
	"strconv"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

var log = clog.NewWithPlugin(thisPlugin)

const thisPlugin = "k8s_crd"

func init() {
	plugin.Register(thisPlugin, setup)
}

func setup(c *caddy.Controller) error {

	gw, err := parse(c)
	if err != nil {
		return plugin.Error(thisPlugin, err)
	}

	gw.Controller, err = RunKubeController(context.Background(), gw)
	if err != nil {
		return plugin.Error(thisPlugin, err)
	}
	gw.ExternalAddrFunc = gw.SelfAddress

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		gw.Next = next
		return gw
	})

	return nil
}

func parseTTL(opt, arg string) (uint32, error) {
	t, err := strconv.Atoi(arg)
	if err != nil {
		return uint32(t), err
	}
	if t < 0 || t > 3600 {
		return uint32(t), fmt.Errorf("%s must be in range [0, 3600]: %d", opt, t)
	}
	return uint32(t), nil
}
func parse(c *caddy.Controller) (*Gateway, error) {
	gw := newGateway()

	for c.Next() {
		zones := c.RemainingArgs()
		gw.Zones = zones

		if len(gw.Zones) == 0 {
			gw.Zones = make([]string, len(c.ServerBlockKeys))
			copy(gw.Zones, c.ServerBlockKeys)
		}

		for i, str := range gw.Zones {
			gw.Zones[i] = plugin.Host(str).Normalize()
		}

		for c.NextBlock() {
			key := c.Val()
			args := c.RemainingArgs()
			if len(args) == 0 {
				return nil, c.ArgErr()
			}
			switch key {
			case "resources":
				gw.updateResources(args)
			case "filter":
				log.Infof("Filter: %+v", args)
				gw.Filter = args[0]
			case "annotation":
				log.Infof("annotation: %+v", args)
				gw.Annotation = args[0]
			case "ttl":
				ttl, err := parseTTL(c.Val(), args[0])
				if err != nil {
					gw.ttlLow = ttl
				}
			case "negttl":
				log.Infof("negTTL: %+v", args[0])
				negttl, err := parseTTL(c.Val(), args[0])
				if err == nil {
					gw.ttlHigh = negttl
				}
			case "apex":
				gw.apex = args[0]
			default:
				return nil, c.Errf("Unknown property '%s'", c.Val())
			}
		}
	}
	return gw, nil

}
