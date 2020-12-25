package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
)

// ParseIPAddresses Parses ip address in a file
func ParseIPAddresses(path string) []string {
	file, err := os.Open(path)
	log.Println("File reading")

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	ipAddresses := []string{}

	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		ipRow := strings.TrimSpace(scanner.Text())
		ip, ipnet, err := net.ParseCIDR(ipRow)

		if err != nil { //nocidr
			ip = net.ParseIP(ipRow)

			if ip == nil {
				log.Fatal("ROW INVALID: " + ipRow)
				continue
			}

			ipAddresses = append(ipAddresses, ip.String())
		} else {
			ips := make([]string, 0)

			for rip := ip.Mask(ipnet.Mask); ipnet.Contains(rip); inc(rip) {
				ips = append(ips, rip.String())
			}

			// remove network address and broadcast address
			lenIPs := len(ips)
			switch {
			case lenIPs < 2:
				ipAddresses = append(ipAddresses, ips...)

			default:
				ipAddresses = append(ipAddresses, ips[1:len(ips)-1]...)
			}

		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return ipAddresses
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
