package agent

import (
	"fmt"
	"log"
	"time"

	"github.com/kingcanfish/ddns-agent/internal/config"
	"github.com/kingcanfish/ddns-agent/internal/detector"
	"github.com/kingcanfish/ddns-agent/internal/dnspod"
	"github.com/kingcanfish/ddns-agent/internal/notifier"
)

type Agent struct {
	config     *config.Config
	ipv4Client *dnspod.Client
	ipv6Client *dnspod.Client
	lanClient  *dnspod.Client
	telegram   *notifier.Telegram
	lastIPv4   string
	lastIPv6   string
	lastLAN    string
}

func New(cfg *config.Config) (*Agent, error) {
	tg := notifier.NewTelegram(cfg.Telegram.BotToken, cfg.Telegram.ChatID)

	if cfg.IPv4URL != "" {
		detector.SetCustomIPv4URL(cfg.IPv4URL)
	}
	if cfg.IPv6URL != "" {
		detector.SetCustomIPv6URL(cfg.IPv6URL)
	}

	agent := &Agent{
		config:   cfg,
		telegram: tg,
	}

	if cfg.IPv4Subdomain != "" {
		agent.ipv4Client = dnspod.NewClient(cfg.Dnspod, cfg.Domain, cfg.IPv4Subdomain)
	}
	if cfg.IPv6Subdomain != "" {
		agent.ipv6Client = dnspod.NewClient(cfg.Dnspod, cfg.Domain, cfg.IPv6Subdomain)
	}
	if cfg.LANSubdomain != "" {
		agent.lanClient = dnspod.NewClient(cfg.Dnspod, cfg.Domain, cfg.LANSubdomain)
	}

	return agent, nil
}

func (a *Agent) Run() {
	log.Printf("Starting DDNS agent for %s", a.config.Domain)
	if a.config.IPv4Subdomain != "" {
		log.Printf("IPv4: %s.%s (A)", a.config.IPv4Subdomain, a.config.Domain)
	}
	if a.config.IPv6Subdomain != "" {
		log.Printf("IPv6: %s.%s (AAAA)", a.config.IPv6Subdomain, a.config.Domain)
	}
	if a.config.LANSubdomain != "" {
		log.Printf("LAN:  %s.%s (A)", a.config.LANSubdomain, a.config.Domain)
	}
	log.Printf("Interval: %ds", a.config.Interval)

	a.checkAndUpdate()

	ticker := time.NewTicker(time.Duration(a.config.Interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		a.checkAndUpdate()
	}
}

func (a *Agent) checkAndUpdate() {
	a.checkIPv4()
	a.checkIPv6()
	a.checkLAN()
}

func (a *Agent) checkIPv4() {
	if a.ipv4Client == nil {
		return
	}

	current, err := detector.GetIPv4()
	if err != nil {
		log.Printf("[IPv4] detect failed: %v", err)
		return
	}

	if current == a.lastIPv4 {
		return
	}

	dnsValue, _ := a.ipv4Client.GetCurrentRecord("A")
	if dnsValue == current {
		log.Printf("[IPv4] %s unchanged", current)
		a.lastIPv4 = current
		return
	}

	log.Printf("[IPv4] updating: %s -> %s", a.lastIPv4, current)
	if err := a.ipv4Client.UpdateRecord("A", current); err != nil {
		log.Printf("[IPv4] update failed: %v", err)
		a.telegram.SendError(fmt.Sprintf("IPv4 update failed: %v", err))
		return
	}

	domain := fmt.Sprintf("%s.%s", a.config.IPv4Subdomain, a.config.Domain)
	a.telegram.SendIPChange(domain, "A (IPv4)", a.lastIPv4, current)
	a.lastIPv4 = current
	log.Printf("[IPv4] updated to %s", current)
}

func (a *Agent) checkIPv6() {
	if a.ipv6Client == nil {
		return
	}

	current, err := detector.GetIPv6()
	if err != nil {
		log.Printf("[IPv6] detect failed: %v", err)
		return
	}

	if current == a.lastIPv6 {
		return
	}

	dnsValue, _ := a.ipv6Client.GetCurrentRecord("AAAA")
	if dnsValue == current {
		log.Printf("[IPv6] %s unchanged", current)
		a.lastIPv6 = current
		return
	}

	log.Printf("[IPv6] updating: %s -> %s", a.lastIPv6, current)
	if err := a.ipv6Client.UpdateRecord("AAAA", current); err != nil {
		log.Printf("[IPv6] update failed: %v", err)
		a.telegram.SendError(fmt.Sprintf("IPv6 update failed: %v", err))
		return
	}

	domain := fmt.Sprintf("%s.%s", a.config.IPv6Subdomain, a.config.Domain)
	a.telegram.SendIPChange(domain, "AAAA (IPv6)", a.lastIPv6, current)
	a.lastIPv6 = current
	log.Printf("[IPv6] updated to %s", current)
}

func (a *Agent) checkLAN() {
	if a.lanClient == nil {
		return
	}

	addresses, err := detector.GetLANAddresses()
	if err != nil {
		log.Printf("[LAN] detect failed: %v", err)
		return
	}

	if len(addresses) == 0 {
		log.Printf("[LAN] no real interface addresses found")
		return
	}

	current := addresses[0]
	if current == a.lastLAN {
		return
	}

	dnsValue, _ := a.lanClient.GetCurrentRecord("A")
	if dnsValue == current {
		log.Printf("[LAN] %s unchanged", current)
		a.lastLAN = current
		return
	}

	log.Printf("[LAN] updating: %s -> %s", a.lastLAN, current)
	if err := a.lanClient.UpdateRecord("A", current); err != nil {
		log.Printf("[LAN] update failed: %v", err)
		a.telegram.SendError(fmt.Sprintf("LAN update failed: %v", err))
		return
	}

	domain := fmt.Sprintf("%s.%s", a.config.LANSubdomain, a.config.Domain)
	a.telegram.SendIPChange(domain, "A (LAN)", a.lastLAN, current)
	a.lastLAN = current
	log.Printf("[LAN] updated to %s", current)
}
