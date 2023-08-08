/*
MIT License

Copyright (c) 2020 The metal-stack Authors.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// frr-monitor compares kernel and zebra routes and eventually restarts frr if some routes are missing
package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/metal-stack/frr-monitor/pkg/frr"
)

// gather kernel and frr routes
func main() {

	// kernelRoutes, err := kernel.GetRoutes()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Kernel Routes")

	// for _, r := range kernelRoutes {
	// 	fmt.Printf("Prefix:%s Nexthop:%s\n", r.Dst, r.Gw)
	// }

	vrfs, err := frr.GetVRFs()
	if err != nil {
		panic(err)
	}
	evpns, err := frr.GetEVPNVNINexthops()
	if err != nil {
		panic(err)
	}
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	for _, vrf := range vrfs {
		if vrf.VrfName == "default" {
			continue
		}
		nexthops := map[string]bool{}
		for _, vr := range vrf.Routes {
			for _, r := range vr {
				for _, nh := range r.Nexthops {
					if nh.Hostname != hostname {
						nexthops[nh.IP] = true
					}
				}
			}
		}
		vni := strings.ReplaceAll(vrf.VrfName, "vrf", "")
		for vrfnh := range nexthops {
			found := false
			for _, evpnnh := range evpns[vni].NextHops {
				if vrfnh == evpnnh.NexthopIp {
					_, err := net.ParseMAC(evpnnh.RouterMac)
					if err == nil {
						found = true
						break
					} else {
						fmt.Printf("VNI:%s VRF:%s Nexthop:%s has invalid mac address %s,%q\n", vni, vrf.VrfName, vrfnh, evpnnh.RouterMac, err)
					}
				}
			}
			if !found {
				fmt.Printf("VNI:%s VRF:%s Nexthop:%s not found in evpn next-hops\n", vni, vrf.VrfName, vrfnh)
			}
		}
	}
}
