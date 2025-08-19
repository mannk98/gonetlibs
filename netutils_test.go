package gonetlibs

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetReportCaller(true)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func TestNetutils(t *testing.T) {
	iface := "cloudbr0"
	ipv4, err := NetGetInterfaceIpv4Addr(iface)
	if err != nil {
		log.Error(err)
	}
	log.Infof("IPv4 of %s is: %s", iface, ipv4)

	if NetIfaceHasIpv4("enp1s0") {
		log.Infof("%s has IPv4", iface)
	} else {
		log.Infof("%s doesn't has IPv4", iface)
	}
	if NetIfaceHasIpv4("vethc025c8d@if10") {
		log.Infof("%s has IPv4", iface)
	} else {
		log.Infof("%s doesn't has IPv4", iface)
	}

	domain := "hitmehardandsoft.site"
	ip, _ := ResolverDomain(domain, true)
	log.Infof("IP address of %s: %v", domain, ip)

	svname := "https://hitmehardandsoft.site.vn"
	checksv := NetCheckConectionToServer(svname, iface)
	if checksv != nil {
		log.Error("Can't not connect to ", svname)
	} else {
		log.Info("Can connect to", svname)
	}

	ifacecheck := "cloudbr0"
	addr, duration, err := Ping("192.168.100.177", ifacecheck)
	if err != nil {
		log.Error(duration, err)
	} else {
		log.Info(addr.IP, duration, err)
	}
	if NetIsOnlinePing(5, 1, ifacecheck) {
		log.Info(ifacecheck, " avaiable to connect to internet.")
	}
}
