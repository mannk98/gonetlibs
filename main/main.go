package main

import (
	"os"

	netutils "github.com/huntelaar112/netutils"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetReportCaller(true)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	iface := "enp1s0"
	ipv4, err := netutils.NetGetInterfaceIpv4Addr(iface)
	if err != nil {
		log.Error(err)
	}
	log.Info("IPv4 of enp1s0 is: ", ipv4)

	if netutils.NetIfaceHasIpv4("enp1s0") {
		log.Info("enp1s0 has IPv4")
	} else {
		log.Info("enp1s0 doesn't has IPv4")
	}
	if netutils.NetIfaceHasIpv4("vethc025c8d@if10") {
		log.Info("enp1s0 has IPv4")
	} else {
		log.Info("enp1s0 doesn't has IPv4")
	}

	ip, _ := netutils.ResolverDomain("stgapi.smartocr.vn", true)
	log.Info("IP address of stgapi.smartocr.vn: ", ip)

	svname := "https://stgapi.smartocr.vn"
	checksv := netutils.NetCheckConectionToServer(svname, "enp1s0")
	if checksv != nil {
		log.Error("Can't not connect to ", svname)
	} else {
		log.Info("Can connect to", svname)
	}

	ifacecheck := "enp1s0"
	addr, duration, err := netutils.Ping("192.168.100.177", ifacecheck)
	if err != nil {
		log.Error(duration, err)
	} else {
		log.Info(addr.IP, duration, err)
	}
	if netutils.NetIsOnlinePing(5, 1, ifacecheck) {
		log.Info(ifacecheck, " avaiable to connect to internet.")
	}
}
