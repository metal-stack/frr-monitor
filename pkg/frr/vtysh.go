package frr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"golang.org/x/exp/slices"
)

type Route struct {
	Valid               bool      `json:"valid"`
	PathFrom            string    `json:"pathFrom"`
	Prefix              string    `json:"prefix"`
	Multipath           bool      `json:"multipath,omitempty"`
	BestPath            bool      `json:"bestpath,omitempty"`
	Selection           string    `json:"selectionReason,omitempty"`
	PrefixLen           int       `json:"prefixLen"`
	Network             string    `json:"network"`
	Version             int       `json:"version"`
	Weight              int       `json:"weight"`
	Metric              *int      `json:"metric,omitempty"`
	PeerID              string    `json:"peerId"`
	Path                string    `json:"path"`
	Origin              string    `json:"origin"`
	AnnounceNexthopSelf bool      `json:"announceNexthopSelf"`
	Nexthops            []Nexthop `json:"nexthops"`
}

type Nexthop struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	AFI      string `json:"afi"`
	Used     bool   `json:"used"`
}

type VRF struct {
	VrfID         int                `json:"vrfId"`
	VrfName       string             `json:"vrfName"`
	TableVersion  int                `json:"tableVersion"`
	RouterID      string             `json:"routerId"`
	DefaultLocPrf int                `json:"defaultLocPrf"`
	LocalAS       int64              `json:"localAS"`
	Routes        map[string][]Route `json:"routes"`
}

type Routes map[string][]Route
type VRFs map[string]VRF

func GetMissingRMAC() ([]string, error) {
	nhs, err := getNexthops()
	if err != nil {
		return nil, err
	}
	evpnnhs, err := getEVPNVNINexthops()
	if err != nil {
		return nil, err
	}

	missing := make(map[string]bool)
	for _, nh := range nhs {
		if !slices.Contains(evpnnhs, nh) {
			missing[nh] = true
		}
	}
	var result []string
	for m := range missing {
		result = append(result, m)
	}
	return result, nil
}

func getNexthops() ([]string, error) {
	output, err := executeVTYSH("bgpd", "show bgp detail json")
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %w", err)
	}

	var vrf VRF
	err = json.Unmarshal(output, &vrf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	var nexthops []string
	for _, vr := range vrf.Routes {
		for _, r := range vr {
			// if r.Metric != nil && *r.Metric == 0 {
			nexthops = append(nexthops, r.Prefix)
			// }
		}
	}

	return nexthops, nil
}

func getEVPNVNINexthops() ([]string, error) {
	output, err := executeVTYSH("zebra", "show evpn next-hops vni all json")
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %w", err)
	}
	var nhs map[string]any
	err = json.Unmarshal(output, &nhs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	var nexthops []string
	for _, v := range nhs {
		value, ok := v.(map[string]any)
		if !ok {
			continue
		}
		for ip := range value {
			nexthops = append(nexthops, ip)
		}
	}

	return nexthops, nil
}

func GetRoutes() (VRFs, error) {
	output, err := executeVTYSH("bgpd", "show bgp vrf all ipv4 unicast json")
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %w", err)
	}
	// fmt.Println(string(output))

	var vrfs VRFs
	err = json.Unmarshal(output, &vrfs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}
	// fmt.Printf("%v\n", vrfs)

	return vrfs, nil
}

func lookupSocketPath(daemon string) (string, error) {
	switch daemon {
	case
		"babeld",
		"bfdd",
		"bgpd",
		"eigrpd",
		"fabricd",
		"isisd",
		"ldpd",
		"nhrpd",
		"ospf6d",
		"ospfd",
		"pbrd",
		"pimd",
		"ripd",
		"ripngd",
		"sharpd",
		"staticd",
		"vrrpd",
		"zebra":
		return fmt.Sprintf("/var/run/frr/%s.vty", daemon), nil
	}
	return "", fmt.Errorf("unknown daemon %s", daemon)
}

func runCmd(socketPath string, cmd string) ([]byte, error) {
	socket, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}
	defer socket.Close()

	cmd = cmd + "\x00"
	_, err = socket.Write([]byte(cmd))
	if err != nil {
		return nil, err
	}

	output, err := bufio.NewReader(socket).ReadBytes('\x00')
	if err != nil {
		return nil, err
	}

	return output[:len(output)-1], nil
}

func executeVTYSH(socket, cmd string) ([]byte, error) {
	socketPath, err := lookupSocketPath(socket)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup socket path: %w", err)
	}

	_, err = runCmd(socketPath, "enable")
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %w", err)
	}
	return runCmd(socketPath, cmd)
}
