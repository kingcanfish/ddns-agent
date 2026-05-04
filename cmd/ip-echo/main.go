package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

func main() {
	ipv4Port := os.Getenv("IPV4_PORT")
	if ipv4Port == "" {
		ipv4Port = "8080"
	}
	ipv6Port := os.Getenv("IPV6_PORT")
	if ipv6Port == "" {
		ipv6Port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIP)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	go func() {
		ln, err := net.Listen("tcp4", ":"+ipv4Port)
		if err != nil {
			log.Fatalf("IPv4 listen failed: %v", err)
		}
		log.Printf("ip-echo IPv4 listening on :%s", ipv4Port)
		log.Fatal(http.Serve(ln, mux))
	}()

	ln, err := net.Listen("tcp6", ":"+ipv6Port)
	if err != nil {
		log.Fatalf("IPv6 listen failed: %v", err)
	}
	log.Printf("ip-echo IPv6 listening on :%s", ipv6Port)
	log.Fatal(http.Serve(ln, mux))
}

func handleIP(w http.ResponseWriter, r *http.Request) {
	ip := extractIP(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Real-IP", ip)
	fmt.Fprint(w, ip)
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
