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
	"log"
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

	// vrfs, err := frr.GetRoutes()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Zebra Routes")
	// for _, vrf := range vrfs {
	// 	for _, vr := range vrf.Routes {
	// 		for _, r := range vr {
	// 			nexthops := []string{}
	// 			for _, nh := range r.Nexthops {
	// 				nexthops = append(nexthops, nh.IP)
	// 			}
	// 			fmt.Printf("Prefix:%s Nexthop:%s\n", r.Prefix, strings.Join(nexthops, ","))
	// 		}
	// 	}
	// }

	missing, err := frr.GetMissingRMAC()
	if err != nil {
		panic(err)
	}

	log.Printf("missing RMAC:%q\n", strings.Join(missing, ","))
}
