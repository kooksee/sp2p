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

// Node represents a host on the network.
// The fields of Node may not be modified.
type Node struct {
	IP  net.IP // len 4 for IPv4 or 16 for IPv6
	UDP uint16 // port numbers
	ID  Hash   // the node's public key

	// Time when the node was added to the table.
	updateAt time.Time
	addr     string
}

// NewNode creates a new node. It is mostly meant to be used for
// testing purposes.
func NewNode(id Hash, ip net.IP, udpPort uint16) *Node {
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}
	return &Node{
		IP:       ip,
		UDP:      udpPort,
		ID:       id,
		updateAt: time.Now(),
	}
}

func (n *Node) Addr() *net.UDPAddr {
	return &net.UDPAddr{IP: n.IP, Port: int(n.UDP)}
}

func (n *Node) AddrString() string {
	if n.addr == "" {
		n.addr = n.Addr().String()
	}
	return n.addr
}

// Incomplete returns true for nodes with no IP address.
func (n *Node) Incomplete() bool {
	return n.IP == nil
}

// checks whether n is a valid complete node.
func (n *Node) validateComplete() error {
	if n.Incomplete() {
		return errors.New("incomplete node")
	}
	if n.UDP == 0 {
		return errors.New("missing UDP port")
	}

	if n.IP.IsMulticast() || n.IP.IsUnspecified() {
		return errors.New("invalid IP (multicast/unspecified)")
	}

	return nil
}

// The string representation of a Node is a URL.
// Please see ParseNode for a description of the format.
func (n *Node) String() string {
	u := url.URL{Scheme: "enode"}
	if n.Incomplete() {
		u.Host = Fmt("%x", n.ID[:])
	} else {
		//u.User = url.User(fmt.Sprintf("%x", n.sha[:]))
		u.User = url.User(Fmt("%x", n.ID[:]))
		u.Host = n.Addr().String()
	}
	return u.String()
}

var incompleteNodeURL = regexp.MustCompile("(?i)^(?:enode://)?([0-9a-f]+)$")

//    enode://<hex node id>@10.3.58.6:30303?discport=30301
//    enode://<hex node id>@10.3.58.6:30303?discport=30301
func ParseNode(rawurl string) (*Node, error) {
	if m := incompleteNodeURL.FindStringSubmatch(rawurl); m != nil {
		id, err := HexID(m[1])
		if err != nil {
			return nil, fmt.Errorf("invalid node ID (%v)", err)
		}
		return NewNode(id, nil, 0), nil
	}
	return parseComplete(rawurl)
}

func parseComplete(rawurl string) (*Node, error) {
	var (
		id               Hash
		ip               net.IP
		tcpPort, udpPort uint64
	)
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "enode" {
		return nil, errors.New("invalid URL scheme, want \"enode\"")
	}
	// Parse the Node ID from the user portion.
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

	return NewNode(id, ip, uint16(udpPort)), nil
}

// MustParseNode parses a node URL. It panics if the URL is not valid.
func MustParseNode(rawUrl string) *Node {
	n, err := ParseNode(rawUrl)
	if err != nil {
		panic(Errs("invalid node URL", err.Error()))
	}
	return n
}
