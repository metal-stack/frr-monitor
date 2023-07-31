package frr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

type (
	BGPSummary struct {
		Ipv4Unicast Ipv4Unicast `json:"ipv4Unicast"`
	}
	Ipv4Unicast struct {
		RouterID    string          `json:"routerId"`
		Peers       map[string]Peer `json:"peers"`
		FailedPeers int             `json:"failedPeers"`
		TotalPeers  int             `json:"totalPeers"`
	}
	Peer struct {
		Hostname                   string `json:"hostname"`
		RemoteAs                   int64  `json:"remoteAs"`
		LocalAs                    int64  `json:"localAs"`
		Version                    int    `json:"version"`
		MsgRcvd                    int    `json:"msgRcvd"`
		MsgSent                    int    `json:"msgSent"`
		TableVersion               int    `json:"tableVersion"`
		Outq                       int    `json:"outq"`
		Inq                        int    `json:"inq"`
		PeerUptime                 string `json:"peerUptime"`
		PeerUptimeMsec             int    `json:"peerUptimeMsec"`
		PeerUptimeEstablishedEpoch int    `json:"peerUptimeEstablishedEpoch"`
		PfxRcd                     int    `json:"pfxRcd"`
		PfxSnt                     int    `json:"pfxSnt"`
		State                      string `json:"state"`
		PeerState                  string `json:"peerState"`
		ConnectionsEstablished     int    `json:"connectionsEstablished"`
		ConnectionsDropped         int    `json:"connectionsDropped"`
		IDType                     string `json:"idType"`
	}

	Routes []Route
	Route  struct {
		Prefix                   string `json:"prefix"`
		PrefixLen                int    `json:"prefixLen"`
		Protocol                 string `json:"protocol"`
		VrfID                    int    `json:"vrfId"`
		VrfName                  string `json:"vrfName"`
		Selected                 bool   `json:"selected"`
		DestSelected             bool   `json:"destSelected"`
		Distance                 int    `json:"distance"`
		Metric                   int    `json:"metric"`
		Installed                bool   `json:"installed"`
		Table                    int    `json:"table"`
		InternalStatus           int    `json:"internalStatus"`
		InternalFlags            int    `json:"internalFlags"`
		InternalNextHopNum       int    `json:"internalNextHopNum"`
		InternalNextHopActiveNum int    `json:"internalNextHopActiveNum"`
		NexthopGroupID           int    `json:"nexthopGroupId"`
		Uptime                   string `json:"uptime"`
		Nexthops                 []struct {
			Flags          int    `json:"flags"`
			Fib            bool   `json:"fib"`
			IP             string `json:"ip"`
			Afi            string `json:"afi"`
			InterfaceIndex int    `json:"interfaceIndex"`
			InterfaceName  string `json:"interfaceName"`
			Active         bool   `json:"active"`
			Weight         int    `json:"weight"`
		} `json:"nexthops"`
		AsPath           string `json:"asPath"`
		Communities      string `json:"communities"`
		LargeCommunities string `json:"largeCommunities"`
	}
)

func GetRoutes() (Routes, error) {
	socketPath, err := lookupSocketPath("bgpd")
	if err != nil {
		return nil, err
	}
	output, err := runCmd(socketPath, "show ip route json")
	if err != nil {
		return nil, err
	}

	var routes Routes
	err = json.Unmarshal(output, &routes)
	if err != nil {
		return nil, err
	}

	return routes, nil
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
