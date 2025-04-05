package gonetlibs

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mannk98/gonetlibs/mdns"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	IfaceIname int = iota
	IfaceMacddr
	IfaceCidr
	IfaceIp4
	IfaceIp6
	IfaceMask
)

type DiscoveryInfo struct {
	Name string
	Host string
	Ip4  string
	Port int
	Info []string
}

var dnslist = []string{
	"1.1.1.1", "1.0.0.1", //clouflare
	//	"208.67.222.222", "208.67.220.220", //opendns server
	"8.8.8.8", "8.8.4.4", //google
	//	"8.26.56.26", "8.20.247.20", //comodo
	//	"9.9.9.9", "149.112.112.112", //quad9
	//	"64.6.64.6", "64.6.65.6"
} // verisign

const (
	// Stolen from https://godoc.org/golang.org/x/net/internal/iana,
	// can't import "internal" packages
	ProtocolICMP = 1
	//ProtocolIPv6ICMP = 58
)

// list all real network iface, include ethernet, wlan, docker network
func NetListAllRealIface() ([]string, error) {
	cmd := exec.Command("ip", "link", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		//("Error running ip command:", err)
		return nil, err
	}

	// Parse the output to extract interface names
	interfaceLines := strings.Split(string(output), "\n")
	interfaces := make([]string, 0)

	for _, line := range interfaceLines {
		fields := strings.Fields(line)
		if len(fields) >= 7 && fields[1] != "lo:" && !strings.Contains(fields[1], "veth") {
			interfaces = append(interfaces, strings.TrimSuffix(fields[1], ":"))
		}
	}

	// Print the list of network interfaces
	//fmt.Println("Network Interfaces:")
	//for _, iface := range interfaces {
	//	fmt.Println(iface)
	//}
	return interfaces, err
}

/* Get first finded IPv4 address of Linux network interface. */
func NetGetInterfaceIpv4Addr(interfaceName string) (addr string, err error) {
	var (
		ief      *net.Interface
		addrs    []net.Addr
		ipv4Addr net.IP
	)
	if ief, err = net.InterfaceByName(interfaceName); err != nil { // get interface
		return
	}
	if addrs, err = ief.Addrs(); err != nil { // get addresses
		return
	}
	for _, addr := range addrs { // get ipv4 address
		if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
			break
		}
	}
	if ipv4Addr == nil {
		if len(addrs) != 0 {
			return "", fmt.Errorf(fmt.Sprintf("There isn't any ipv4 on interface %s\n", interfaceName))
		} else {
			return "", fmt.Errorf(fmt.Sprintf("There isn't any ip on interface %s\n", interfaceName))
		}
	}
	return ipv4Addr.String(), nil
}

/* Check if Network Interface has any IPv4 address. */
func NetIfaceHasIpv4(interfaceName string) bool {
	if _, err := NetGetInterfaceIpv4Addr(interfaceName); err == nil {
		return true
	}
	return false
}

/* Convert Domain to IP */
func ResolverDomain(domain string, debugflag ...bool) (addrs []string, err error) {
	if addr := net.ParseIP(domain); addr != nil {
		return []string{domain}, nil
	}

	r := &net.Resolver{
		PreferGo: true,
		Dial:     nil,
	}

	for i := 0; i < len(dnslist); i++ {
		for _, pro := range []string{"udp", "tcp"} {
			/* setup a custom dial for Resolver, return connection to host (udp or tcp) */
			r.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Millisecond * time.Duration(5000),
				}
				return d.DialContext(ctx, pro, dnslist[i]+":53")
			}

			/* start lookup using custom resolver, if success, return. */
			if addrs, err = r.LookupHost(context.Background(), domain); err == nil {
				return addrs, err
			}

			if len(debugflag) != 0 && debugflag[0] {
				log.Errorf("\nCan not used dns server %s for finding %s\n", dnslist[i], domain)
			}
		}
	}
	return net.LookupHost(domain) //system lockup if custome resolver fail.
}

func ResolverDomain2Ip4(domain string, debugflag ...bool) (addr string, err error) {
	if addrs, err := ResolverDomain(domain, debugflag...); err == nil {
		for _, v := range addrs {
			if strings.Contains(v, ".") {
				return v, nil
			}
		}
		return "", fmt.Errorf("there is not ipv4")
	} else {
		return "", err
	}
}

/*
	Check connection to http/https server

return nil if cant connect to server through interface
*/
func NetCheckConectionToServer(domain string, ifacenames ...string) error {
	tcpAddr := &net.TCPAddr{}

	if len(ifacenames) != 0 {
		ip4add, err := NetGetInterfaceIpv4Addr(ifacenames[0])
		if err != nil || len(ip4add) == 0 {
			return err
		} else {
			tcpAddr.IP = net.ParseIP(ip4add)
		}
	} else {
		tcpAddr = nil
	}

	d := net.Dialer{LocalAddr: tcpAddr, Timeout: time.Millisecond * 2000}

	if !strings.Contains(domain, "://") {
		domain = "http://" + domain
	}
	u, err := url.Parse(domain)
	if err != nil {
		return err
	}
	port := "80"
	if u.Scheme == "https" {
		port = "443"
	}

	host := u.Host
	if thost, tport, _ := net.SplitHostPort(u.Host); len(thost) != 0 {
		port = tport
		host = thost
	}
	ip4, err := ResolverDomain2Ip4(host)
	if err != nil {
		//		log.Error(err, host)

		return err
	}
	if conn, err := d.Dial("tcp", ip4+":"+port); err != nil {
		//		log.Error(err)
		return err
	} else {
		conn.Close()
		return nil
	}
}

/* Check if server is alive, timeout check is 666ms */
func ServerIsLive(domain string, ifacenames ...string) bool {
	tcpAddr := &net.TCPAddr{}

	if len(ifacenames) != 0 {
		ip4add, err := NetGetInterfaceIpv4Addr(ifacenames[0])
		if err != nil || len(ip4add) == 0 {
			return false
		} else {
			tcpAddr.IP = net.ParseIP(ip4add)
		}
	} else {
		tcpAddr = nil
	}

	d := net.Dialer{LocalAddr: tcpAddr, Timeout: time.Millisecond * 666}

	if !strings.Contains(domain, "://") {
		domain = "http://" + domain
	}
	u, err := url.Parse(domain)
	if err != nil {
		return false
	}
	port := "80"
	if u.Scheme == "https" {
		port = "443"
	}

	host := u.Host
	if thost, tport, _ := net.SplitHostPort(u.Host); len(thost) != 0 {
		port = tport
		host = thost
	}
	ip4, err := ResolverDomain2Ip4(host)
	if err != nil {
		//		log.Error(err, host)

		return false
	}
	if conn, err := d.Dial("tcp", ip4+":"+port); err != nil {
		//		log.Error(err)
		return false
	} else {
		conn.Close()
		return true
	}
}

/* Check if host machine have internet (check http connection to DNS server ) */
func NetIsOnlineTcp(times, intervalsecs int, ifacenames ...string) bool {
	ifacename := ""
	if len(ifacenames) != 0 {
		ifacename = ifacenames[0]
	}
	//	timeout := time.Millisecond * 500
	//	if sutils.StringContainsI(ifacename, "ppp") {
	//		timeout = time.Millisecond * 3000
	//	}
	numDnsTest := len(dnslist)
	//	if numDnsTest >= 4 {
	//		numDnsTest = 4
	//	}
	ttk := time.NewTicker(time.Second * time.Duration(intervalsecs))
	for i1 := 0; i1 < times; i1++ {
		for i := 0; i < numDnsTest; i++ {
			//			log.Warn("Ping interface ", ifacename, dnslist[i])
			if ServerIsLive(dnslist[i], ifacename) {
				return true
			} else {
				//				log.Errorf("Error to use iface %s to test dns server: %s\n", ifacename, dnslist[i])
				continue
			}
		}
		if times > 1 {
			<-ttk.C
		}
	}
	return false
}

/* Check if host machine have internet (check by ping(imcp) to dns servers) */
func NetIsOnlinePing(times, intervalsecs int, ifacenames ...string) bool {
	ifacename := ""
	if len(ifacenames) != 0 {
		ifacename = ifacenames[0]
	}
	timeout := time.Millisecond * 500
	//	if sutils.StringContainsI(ifacename, "ppp") {
	//		timeout = time.Millisecond * 3000
	//	}
	numDnsTest := len(dnslist)
	//	if numDnsTest >= 4 {
	//		numDnsTest = 4
	//	}
	ttk := time.NewTicker(time.Second * time.Duration(intervalsecs))
	for i1 := 0; i1 < times; i1++ {
		for i := 0; i < numDnsTest; i++ {
			//			log.Warn("Ping interface ", ifacename)
			if _, _, err := Ping(dnslist[i], ifacename, timeout); err != nil {
				//				if sutils.StringContainsI(ifacename, "ppp") {
				//				log.Errorf("Error to use iface %s to test dns server: %s\n%s\n", ifacename, dnslist[i], err.Error())
				//				}
				continue
			} else {
				//				log.Warnf("Use iface %s to test dns server: %s\n", ifacename, dnslist[i])
				return true
			}
		}
		if times > 1 {
			<-ttk.C
		}
	}
	return false
}

func Ping(addr, iface string, timeouts ...time.Duration) (*net.IPAddr, time.Duration, error) {
	// Start listening for icmp replies
	timeout := time.Millisecond * 1000
	if len(timeouts) != 0 {
		timeout = timeouts[0]
	}
	listenAddr := "0.0.0.0"
	if len(iface) != 0 {
		// if sutils.StringContainsI(iface, "ppp") {
		//			defer sutils.TimeTrack(time.Now())
		// }
		if ip4, err := NetGetInterfaceIpv4Addr(iface); err == nil {
			//			fmt.Printf("\nip's%s is %s\n", iface, ip4)
			listenAddr = ip4
		} else {
			return nil, 0, fmt.Errorf("iface %s don't have ipv4 address.", iface)
		}
	}
	//	c, err := icmp.ListenPacket("ip4:1991", listenAddr+":1991")
	c := new(icmp.PacketConn)
	var err error
	c, err = icmp.ListenPacket("ip4:icmp", listenAddr)
	/* 	if sutils.GOOS != "windows" {
	   		c, err = icmp.ListenPacket("udp4", listenAddr)
	   	} else {
	   		c, err = icmp.ListenPacket("ip4:icmp", listenAddr)
	   	} */

	if err != nil {
		return nil, 0, err
	}
	defer c.Close()

	// Resolve any DNS (if used) and get the real IP of the target
	dstip4, err := ResolverDomain2Ip4(addr)
	if err != nil {
		//		panic(err)
		return nil, 0, err
	}

	dst, err := net.ResolveIPAddr("ip4", dstip4)
	if err != nil {
		//		panic(err)
		return nil, 0, err
	}
	// Make a new ICMP message
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1, //<< uint(seq), // TODO
			Data: []byte(""),
		},
	}
	b, err := m.Marshal(nil)
	if err != nil {
		return dst, 0, err
	}

	// Send it
	start := time.Now()
	n, err := c.WriteTo(b, dst)
	if err != nil {
		return dst, 0, err
	} else if n != len(b) {
		return dst, 0, fmt.Errorf("got %v; want %v", n, len(b))
	}

	// Wait for a reply
	reply := make([]byte, 1500)
	err = c.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return dst, 0, err
	}
	n, peer, err := c.ReadFrom(reply)
	if err != nil {
		return dst, 0, err
	}
	duration := time.Since(start)

	// Pack it up boys, we're done here
	rm, err := icmp.ParseMessage(ProtocolICMP, reply[:n])
	if err != nil {
		return dst, 0, err
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		return dst, duration, nil
	default:
		return dst, 0, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
	}
}

// id is instance in mdns
func NetInitDiscoveryServer(ipService interface{}, serviceport interface{}, id, serviceName string, info interface{}, ifaceName ...interface{}) (s *mdns.Server, err error) {

	if len(serviceName) == 0 {
		serviceName = "_signage._tcp"
	}
	//(instance, service, domain, hostName string, port int, ips []net.IP, txt []string)
	service, err := mdns.NewMDNSService(id, serviceName, "", "", serviceport, ipService, info)

	if err != nil {
		log.Error("Cannot create config mdnsServer", err)
		return nil, err
	}
	var iface *net.Interface

	if len(ifaceName) != 0 {
		switch v := ifaceName[0].(type) {
		case string:
			if ief, err := net.InterfaceByName(v); err == nil { // get interface
				iface = ief
			}
		case net.Interface:
			iface = &v
		case *net.Interface:
			iface = v
		default:
		}
	}
	// Create the mDNS server, defer shutdown
	if s, err = mdns.NewServer(&mdns.Config{Zone: service, Iface: iface}); err != nil {
		log.Error("Cannot start mdnsServer", err)
		return nil, err
	} else {
		//		log.Print(s.Config.Zone.(*mdns.MDNSService).Port)
		return s, nil
	}
	//	defer s.Shutdown()
}

func NetDiscoveryQueryServiceEntry(serviceName, domain string, timeout time.Duration, ifaceNames ...string) []*mdns.ServiceEntry {
	serviceInfo := make([]*mdns.ServiceEntry, 0)
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 16)
	mutex := &sync.Mutex{}
	go func() {
		for entry := range entriesCh {
			// fmt.Printf("Got new signage entry: %v\n", entry)
			mutex.Lock()
			serviceInfo = append(serviceInfo, entry)
			mutex.Unlock()
		}
	}()

	if len(serviceName) == 0 {
		serviceName = "_signage._tcp"
	}
	params := mdns.DefaultParams(serviceName)
	params.WantUnicastResponse = true
	if len(ifaceNames) != 0 {
		if ief, err := net.InterfaceByName(ifaceNames[0]); err == nil { // get interface
			//			log.Info("NetDiscoveryQuery on interface ", ifaceNames[0])
			params.Interface = ief
		} else {
			return serviceInfo
		}
	} else {
		log.Error("NetDiscoveryQuery on default interface")
	}
	params.Domain = ""
	params.Entries = entriesCh
	if timeout == 0 {
		timeout = time.Millisecond * 600
	}
	params.Timeout = timeout
	params.Domain = domain
	if err := mdns.Query(params); err != nil {
		log.Error("mdns.Query err:", err)
	}
	// Start the lookup
	//	mdns.Lookup("_signage._tcp", entriesCh)
	close(entriesCh)
	return serviceInfo
}
