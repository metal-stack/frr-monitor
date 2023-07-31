package kernel

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

func GetRoutes() ([]string, error) {

	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		routesv4, err := netlink.RouteList(link, netlink.FAMILY_V4)
		if err != nil {
			return nil, err
		}

		for _, route := range routesv4 {
			fmt.Printf("Link: %q Route:%v\n", link.Attrs().Name, route)
		}
	}

	return nil, nil
}
