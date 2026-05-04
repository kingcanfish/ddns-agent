package dnspod

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const apiBase = "https://dnsapi.cn"

type Client struct {
	token     string
	domain    string
	subdomain string
	client    *http.Client
}

type apiResponse struct {
	Status struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		CreatedAt string `json:"created_at"`
	} `json:"status"`
	Records []record `json:"records"`
	Record  struct {
		ID string `json:"id"`
	} `json:"record"`
}

type record struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Line    string `json:"line"`
	LineID  string `json:"line_id"`
	Type    string `json:"type"`
	TTL     string `json:"ttl"`
	Value   string `json:"value"`
	Enabled string `json:"enabled"`
}

func NewClient(token, domain, subdomain string) *Client {
	return &Client{
		token:     token,
		domain:    domain,
		subdomain: subdomain,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) GetCurrentRecord(recordType string) (string, error) {
	params := url.Values{}
	params.Set("login_token", c.token)
	params.Set("format", "json")
	params.Set("domain", c.domain)
	params.Set("sub_domain", c.subdomain)
	params.Set("record_type", recordType)
	params.Set("offset", "0")
	params.Set("length", "1")

	data, err := c.request("Record.List", params)
	if err != nil {
		return "", err
	}

	var resp apiResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if resp.Status.Code != "1" {
		return "", fmt.Errorf("api error: %s", resp.Status.Message)
	}

	if len(resp.Records) == 0 {
		return "", nil
	}

	return resp.Records[0].Value, nil
}

func (c *Client) UpdateRecord(recordType, value string) error {
	recordID, err := c.findRecord(recordType)
	if err != nil || recordID == "" {
		log.Printf("Record not found, creating new one")
		return c.createRecord(recordType, value)
	}

	return c.modifyRecord(recordID, recordType, value)
}

func (c *Client) findRecord(recordType string) (string, error) {
	params := url.Values{}
	params.Set("login_token", c.token)
	params.Set("format", "json")
	params.Set("domain", c.domain)
	params.Set("sub_domain", c.subdomain)
	params.Set("record_type", recordType)
	params.Set("offset", "0")
	params.Set("length", "1")

	data, err := c.request("Record.List", params)
	if err != nil {
		return "", err
	}

	var resp apiResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if resp.Status.Code != "1" {
		return "", fmt.Errorf("api error: %s", resp.Status.Message)
	}

	if len(resp.Records) == 0 {
		return "", nil
	}

	return resp.Records[0].ID, nil
}

func (c *Client) createRecord(recordType, value string) error {
	params := url.Values{}
	params.Set("login_token", c.token)
	params.Set("format", "json")
	params.Set("domain", c.domain)
	params.Set("sub_domain", c.subdomain)
	params.Set("record_type", recordType)
	params.Set("record_line", "默认")
	params.Set("value", value)
	params.Set("ttl", "600")

	data, err := c.request("Record.Create", params)
	if err != nil {
		return err
	}

	var resp apiResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if resp.Status.Code != "1" {
		return fmt.Errorf("api error: %s", resp.Status.Message)
	}

	log.Printf("Created record ID: %s", resp.Record.ID)
	return nil
}

func (c *Client) modifyRecord(recordID, recordType, value string) error {
	params := url.Values{}
	params.Set("login_token", c.token)
	params.Set("format", "json")
	params.Set("domain", c.domain)
	params.Set("record_id", recordID)
	params.Set("sub_domain", c.subdomain)
	params.Set("record_type", recordType)
	params.Set("record_line", "默认")
	params.Set("value", value)
	params.Set("ttl", "600")

	data, err := c.request("Record.Modify", params)
	if err != nil {
		return err
	}

	var resp apiResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if resp.Status.Code != "1" {
		return fmt.Errorf("api error: %s", resp.Status.Message)
	}

	log.Printf("Modified record: %s", recordID)
	return nil
}

func (c *Client) request(action string, params url.Values) ([]byte, error) {
	apiURL := fmt.Sprintf("%s/%s", apiBase, action)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", action, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = strings.TrimSpace
}
