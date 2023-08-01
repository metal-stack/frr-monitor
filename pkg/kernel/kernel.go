package kernel

import (
	"errors"
	"fmt"

	"github.com/vishvananda/netlink"
)

func GetRoutes() ([]netlink.Route, error) {

	routes, err := netlink.RouteListFiltered(netlink.FAMILY_V4, &netlink.Route{Table: 0}, netlink.RT_FILTER_TABLE)

	if err != nil {
		fmt.Printf("Failed to get routes for VRF table %d: %v", 0, err)
	}

	return routes, errors.Join(err)
}
