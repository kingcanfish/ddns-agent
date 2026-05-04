package detector

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var defaultIPv4Services = []string{
	"https://ip.sb",
	"https://api.ipify.org",
	"https://icanhazip.com",
}

var defaultIPv6Services = []string{
	"https://ip.sb",
	"https://api6.ipify.org",
	"https://icanhazip.com",
}

var customIPv4URL string
var customIPv6URL string

func SetCustomIPv4URL(url string) {
	customIPv4URL = strings.TrimRight(url, "/")
}

func SetCustomIPv6URL(url string) {
	customIPv6URL = strings.TrimRight(url, "/")
}

func GetIPv4() (string, error) {
	if customIPv4URL != "" {
		ip, err := fetchIP(customIPv4URL)
		if err == nil && isValidIPv4(ip) {
			return ip, nil
		}
		return "", fmt.Errorf("custom IPv4 service failed: %w", err)
	}

	for _, service := range defaultIPv4Services {
		ip, err := fetchIP(service)
		if err == nil && isValidIPv4(ip) {
			return ip, nil
		}
	}
	return "", fmt.Errorf("failed to get IPv4 from all services")
}

func GetIPv6() (string, error) {
	if customIPv6URL != "" {
		ip, err := fetchIP(customIPv6URL)
		if err == nil && isValidIPv6(ip) {
			return ip, nil
		}
		return "", fmt.Errorf("custom IPv6 service failed: %w", err)
	}

	for _, service := range defaultIPv6Services {
		ip, err := fetchIP(service)
		if err == nil && isValidIPv6(ip) {
			return ip, nil
		}
	}
	return "", fmt.Errorf("failed to get IPv6 from all services")
}

func fetchIP(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

func isValidIPv4(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

func isValidIPv6(ip string) bool {
	if strings.Count(ip, ":") < 2 {
		return false
	}
	if strings.Contains(ip, ".") {
		return false
	}
	return true
}
