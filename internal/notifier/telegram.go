package notifier

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Telegram struct {
	BotToken string
	ChatID   string
}

func NewTelegram(botToken, chatID string) *Telegram {
	return &Telegram{
		BotToken: botToken,
		ChatID:   chatID,
	}
}

func (t *Telegram) SendIPChange(domain, recordType, oldIP, newIP string) error {
	if t.BotToken == "" || t.ChatID == "" {
		log.Println("Telegram notification skipped: bot token or chat ID not configured")
		return nil
	}

	message := fmt.Sprintf("🔄 DDNS IP 变更通知\n域名: %s\n类型: %s\n旧 IP: %s\n新 IP: %s\n时间: %s",
		domain, recordType, oldIP, newIP, time.Now().Format("2006-01-02 15:04:05"))

	return t.send(message)
}

func (t *Telegram) SendError(errMsg string) error {
	if t.BotToken == "" || t.ChatID == "" {
		return nil
	}

	message := fmt.Sprintf("❌ DDNS 错误\n错误: %s\n时间: %s", errMsg, time.Now().Format("2006-01-02 15:04:05"))

	return t.send(message)
}

func (t *Telegram) send(text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)

	params := url.Values{}
	params.Set("chat_id", t.ChatID)
	params.Set("text", text)
	params.Set("parse_mode", "HTML")

	resp, err := http.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf("send telegram message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: %s", string(body))
	}

	log.Printf("Telegram notification sent successfully")
	return nil
}
