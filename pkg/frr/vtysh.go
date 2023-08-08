package frr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

type Route struct {
	Valid               bool      `json:"valid"`
	PathFrom            string    `json:"pathFrom"`
	Prefix              string    `json:"prefix"`
	PrefixLen           int       `json:"prefixLen"`
	Network             string    `json:"network"`
	Version             int       `json:"version"`
	Weight              int       `json:"weight"`
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

type NextHopInfo struct {
	NexthopIp string `json:"nexthopIp"`
	RouterMac string `json:"routerMac"`
}

type VRFInfo struct {
	NumNextHops int                    `json:"numNextHops"`
	NextHops    map[string]NextHopInfo `json:"nexthops"`
}

type Routes map[string][]Route
type VRFs map[string]VRF

func GetVRFs() (VRFs, error) {
	socketPath, err := lookupSocketPath("bgpd")
	if err != nil {
		return nil, fmt.Errorf("failed to lookup socket path: %w", err)
	}

	_, err = runCmd(socketPath, "enable")
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %w", err)
	}
	output, err := runCmd(socketPath, "show bgp vrf all ipv4 unicast json")
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %w", err)
	}

	var vrfs VRFs
	err = json.Unmarshal(output, &vrfs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return vrfs, nil
}

func GetEVPNVNINexthops() (map[string]VRFInfo, error) {
	output, err := executeVTYSH("zebra", "show evpn next-hops vni all json")
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %w", err)
	}
	vrfs := map[string]VRFInfo{}

	var vrfData map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(output), &vrfData); err != nil {
		return nil, err
	}

	for vrfID, vrfInfo := range vrfData {
		numNextHops := int(vrfInfo["numNextHops"].(float64))
		nextHops := map[string]NextHopInfo{}
		for ip, hop := range vrfInfo {
			if ip != "numNextHops" {
				nextHop := hop.(map[string]interface{})
				nextHops[ip] = NextHopInfo{
					NexthopIp: nextHop["nexthopIp"].(string),
					RouterMac: nextHop["routerMac"].(string),
				}
			}
		}
		vrfs[vrfID] = VRFInfo{
			NumNextHops: numNextHops,
			NextHops:    nextHops,
		}
	}

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
