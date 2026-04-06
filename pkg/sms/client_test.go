package sms

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	cfg := Config{
		AccessKeyID:     "test-id",
		AccessKeySecret: "test-secret",
		SignName:        "TestSign",
		TemplateCode:    "SMS_123456",
	}
	client, err := NewClient(cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "dysmsapi.aliyuncs.com", client.cfg.Endpoint)
}

func TestSendCodeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("Action"), "SendSms")
		assert.Equal(t, "13800138000", r.URL.Query().Get("PhoneNumbers"))
		assert.Equal(t, "SMS_123456", r.URL.Query().Get("TemplateCode"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"Code":    "OK",
			"Message": "success",
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		AccessKeyID:     "test-id",
		AccessKeySecret: "test-secret",
		SignName:        "TestSign",
		TemplateCode:    "SMS_123456",
		Endpoint:        server.URL[7:],
	})

	err := client.SendCode(context.Background(), "13800138000", "123456")
	require.NoError(t, err)
}

func TestSendCodeAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"Code":    "isv.BUSINESS_LIMIT_CONTROL",
			"Message": "trigger frequency limit",
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		AccessKeyID:     "test-id",
		AccessKeySecret: "test-secret",
		SignName:        "TestSign",
		TemplateCode:    "SMS_123456",
		Endpoint:        server.URL[7:],
	})

	err := client.SendCode(context.Background(), "13800138000", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "isv.BUSINESS_LIMIT_CONTROL")
}
