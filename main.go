package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

func isIPInRange(ipStr, cidr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("Invalid IP address: %s", ipStr)
	}

	_, ipNet, ipNetErr := net.ParseCIDR(cidr)
	if ipNetErr != nil {
		return false, fmt.Errorf("Invalid CIDR range: %s", cidr)
	}

	return ipNet.Contains(ip), nil
}

func main() {
	var logFilePath, cidrFilePath string
	var invertMatch bool

	flag.StringVar(&logFilePath, "log", "/dev/stdin", "Path to the log file")
	flag.StringVar(&cidrFilePath, "cidr", "", "Path to the CIDR file")
	flag.BoolVar(&invertMatch, "invert", false, "Invert the match (print IP addresses not in the provided CIDR ranges)")

	flag.Parse()

	if logFilePath == "" || cidrFilePath == "" {
		fmt.Println("Usage: greip -log <logfile> -cidr <cidrfile> [-invert]")
		os.Exit(1)
	}

	// Read CIDRs from file
	cidrFile, err := os.Open(cidrFilePath)
	if err != nil {
		fmt.Printf("Error opening CIDR file: %v\n", err)
		os.Exit(1)
	}
	defer cidrFile.Close()

	var cidrs []string
	scanner := bufio.NewScanner(cidrFile)
	for scanner.Scan() {
		cidr := strings.TrimSpace(scanner.Text())
		if cidr != "" {
			cidrs = append(cidrs, cidr)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading CIDR file: %v\n", err)
		os.Exit(1)
	}

	// Updated IPv4 regex
	ipRegex := regexp.MustCompile(`^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.){3}(25[0-5]|(2[0-4]|1\d|[1-9]|)\d)$`)

	// Read and process log file
	logFile, err := os.Open(logFilePath)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	scanner = bufio.NewScanner(logFile)
	for scanner.Scan() {
		line := scanner.Text()
		ips := ipRegex.FindAllString(line, -1)
		if ips != nil {
			for _, ip := range ips {
				matched := false
				for _, cidr := range cidrs {
					inRange, err := isIPInRange(ip, cidr)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						continue
					}

					if inRange {
						matched = true
						break
					}
				}

				if invertMatch && !matched {
					//fmt.Printf("%s is not in any of the CIDR ranges\n", ip)
					fmt.Printf("%s\n", ip)
				} else if !invertMatch && matched {
					//fmt.Printf("%s is in at least one of the CIDR ranges\n", ip)
					fmt.Printf("%s\n", ip)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
		os.Exit(1)
	}
}
