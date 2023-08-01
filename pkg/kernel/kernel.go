package kernel

import (
	"errors"

	"github.com/vishvananda/netlink"
)

func GetRoutes() ([]netlink.Route, error) {

	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	var routes []netlink.Route
	var errs []error
	for _, link := range links {
		routesv4, err := netlink.RouteList(link, netlink.FAMILY_V4)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		routes = append(routes, routesv4...)
		routesv6, err := netlink.RouteList(link, netlink.FAMILY_V6)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		routes = append(routes, routesv6...)
	}

	return routes, errors.Join(errs...)
}
