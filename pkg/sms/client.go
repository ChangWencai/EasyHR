package sms

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
	Endpoint        string
	TestMode        bool
}

type Client struct {
	cfg    Config
	client *resty.Client
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "dysmsapi.aliyuncs.com"
	}
	return &Client{
		cfg:    cfg,
		client: resty.New().SetTimeout(10 * time.Second),
	}, nil
}

func (c *Client) IsTestMode() bool {
	return c.cfg.TestMode
}

func (c *Client) SendCode(ctx context.Context, phone, code string) error {
	if c.cfg.TestMode {
		return nil
	}
	params := map[string]string{
		"AccessKeyId":      c.cfg.AccessKeyID,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     phone,
		"RegionId":         "cn-hangzhou",
		"SignName":         c.cfg.SignName,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"SignatureVersion": "1.0",
		"TemplateCode":     c.cfg.TemplateCode,
		"TemplateParam":    fmt.Sprintf(`{"code":"%s"}`, code),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}

	sign := signRequest(params, c.cfg.AccessKeySecret)
	params["Signature"] = sign

	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}

	scheme := "https"
	if strings.HasPrefix(c.cfg.Endpoint, "http://") || strings.HasPrefix(c.cfg.Endpoint, "localhost") || strings.HasPrefix(c.cfg.Endpoint, "127.0.0.1") {
		scheme = "http"
	}
	endpoint := c.cfg.Endpoint
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = scheme + "://" + endpoint
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryString(query.Encode()).
		Get(endpoint)
	if err != nil {
		return fmt.Errorf("send sms: %w", err)
	}

	var result struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	if result.Code != "OK" {
		return fmt.Errorf("sms api error: %s - %s", result.Code, result.Message)
	}
	return nil
}

// SendTemplateMessage 发送自定义模板短信
func (c *Client) SendTemplateMessage(ctx context.Context, phone, templateCode, templateParam string) error {
	if c.cfg.TestMode {
		return nil
	}
	params := map[string]string{
		"AccessKeyId":      c.cfg.AccessKeyID,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     phone,
		"RegionId":         "cn-hangzhou",
		"SignName":         c.cfg.SignName,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"SignatureVersion": "1.0",
		"TemplateCode":     templateCode,
		"TemplateParam":    templateParam,
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}

	sign := signRequest(params, c.cfg.AccessKeySecret)
	params["Signature"] = sign

	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}

	scheme := "https"
	if strings.HasPrefix(c.cfg.Endpoint, "http://") || strings.HasPrefix(c.cfg.Endpoint, "localhost") || strings.HasPrefix(c.cfg.Endpoint, "127.0.0.1") {
		scheme = "http"
	}
	endpoint := c.cfg.Endpoint
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = scheme + "://" + endpoint
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryString(query.Encode()).
		Get(endpoint)
	if err != nil {
		return fmt.Errorf("send template sms: %w", err)
	}

	var result struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	if result.Code != "OK" {
		return fmt.Errorf("sms api error: %s - %s", result.Code, result.Message)
	}
	return nil
}

func signRequest(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sorted string
	for i, k := range keys {
		if i > 0 {
			sorted += "&"
		}
		sorted += url.QueryEscape(k) + "=" + url.QueryEscape(params[k])
	}

	stringToSign := "GET&%2F&" + url.QueryEscape(sorted)
	mac := hmac.New(sha1.New, []byte(secret+"&"))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
