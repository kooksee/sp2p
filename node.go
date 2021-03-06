package sp2p

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// node represents a host on the network.
// The fields of node may not be modified.
type node struct {
	IP   net.IP // len 4 for IPv4 or 16 for IPv6
	Port uint16 // port numbers
	ID   Hash   // the node's public key

	// Time when the node was added to the table.
	updateAt   time.Time
	addr       string
	udpAddr    *net.UDPAddr
	nodeString string
}

// Newnode creates a new node. It is mostly meant to be used for
// testing purposes.
func newNode(id Hash, ip net.IP, udpPort uint16) *node {
	n := &node{
		IP:       ip,
		Port:     udpPort,
		ID:       id,
		addr:     fmt.Sprintf("%s:%d", ip.String(), udpPort),
		updateAt: time.Now(),
		udpAddr:  &net.UDPAddr{IP: ip, Port: int(udpPort)},
	}
	n.nodeString = n.string()

	return n
}

func (n *node) adds() *net.UDPAddr {
	return n.udpAddr
}

func (n *node) addrString() string {
	return n.addr
}

// Incomplete returns true for nodes with no IP address.
func (n *node) incomplete() bool {
	return n.IP == nil
}

// checks whether n is a valid complete node.
func (n *node) validateComplete() error {
	if n.incomplete() {
		return errors.New("incomplete node")
	}
	if n.Port == 0 {
		return errors.New("missing UDP port")
	}

	if n.IP.IsMulticast() || n.IP.IsUnspecified() {
		return errors.New("invalid IP (multicast/unspecified)")
	}

	return nil
}

// The string representation of a node is a URL.
// Please see Parsenode for a description of the format.
func (n *node) string() string {
	if n.nodeString != "" {
		return n.nodeString
	}

	u := url.URL{Scheme: "sp2p"}
	if n.incomplete() {
		u.Host = f("%x", n.ID[:])
	} else {
		//u.User = url.User(fmt.Sprintf("%x", n.sha[:]))
		u.User = url.User(f("%x", n.ID[:]))
		u.Host = n.addrString()
	}
	n.nodeString = u.String()

	return n.nodeString
}

var incompletenodeURL = regexp.MustCompile("(?i)^(?:sp2p://)?([0-9a-f]+)$")

//    sp2p://<hex node id>@10.3.58.6:30303?discport=30301
//    sp2p://<hex node id>@10.3.58.6:30303?discport=30301
func NodeParse(rawurl string) (*node, error) {
	if m := incompletenodeURL.FindStringSubmatch(rawurl); m != nil {
		id, err := HexID(m[1])
		if err != nil {
			return nil, fmt.Errorf("invalid node ID (%v)", err)
		}
		return newNode(id, nil, 0), nil
	}
	return parseComplete(rawurl)
}

func parseComplete(rawurl string) (*node, error) {
	var (
		id               Hash
		ip               net.IP
		tcpPort, udpPort uint64
	)
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "sp2p" {
		return nil, errors.New("invalid URL scheme, want \"sp2p\"")
	}
	// Parse the node ID from the user portion.
	if u.User == nil {
		return nil, errors.New("does not contain node ID")
	}
	if id, err = HexID(u.User.String()); err != nil {
		return nil, fmt.Errorf("invalid node ID (%v)", err)
	}
	// Parse the IP address.
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid host: %v", err)
	}
	if ip = net.ParseIP(host); ip == nil {
		return nil, errors.New("invalid IP address")
	}
	// Ensure the IP is 4 bytes long for IPv4 addresses.
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}
	// Parse the port numbers.
	if tcpPort, err = strconv.ParseUint(port, 10, 16); err != nil {
		return nil, errors.New("invalid port")
	}
	udpPort = tcpPort
	qv := u.Query()
	if qv.Get("discport") != "" {
		udpPort, err = strconv.ParseUint(qv.Get("discport"), 10, 16)
		if err != nil {
			return nil, errors.New("invalid discport in query")
		}
	}

	return newNode(id, ip, uint16(udpPort)), nil
}

// MustNodeParse parses a node URL. It panics if the URL is not valid.
func MustNodeParse(rawUrl string) *node {
	n, err := NodeParse(rawUrl)
	if err != nil {
		panic(errs("invalid node URL", err.Error()))
	}
	return n
}
