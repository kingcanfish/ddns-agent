package detector

import (
	"fmt"
	"net"
	"strings"
)

var virtualPrefixes = []string{
	"docker",
	"veth",
	"br-",
	"virbr",
	"tun",
	"tap",
	"wg",
	"lo",
	"vmnet",
	"vboxnet",
}

func GetLANAddresses() ([]string, error) {
	var addresses []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("get network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		if isVirtualInterface(iface.Name) {
			continue
		}

		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}

			addresses = append(addresses, ip.String())
		}
	}

	return addresses, nil
}

func isVirtualInterface(name string) bool {
	for _, prefix := range virtualPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
