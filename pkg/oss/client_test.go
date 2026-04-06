package oss

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFileType(t *testing.T) {
	assert.True(t, ValidateFileType("image/jpeg"))
	assert.True(t, ValidateFileType("image/png"))
	assert.True(t, ValidateFileType("application/pdf"))
	assert.False(t, ValidateFileType("application/zip"))
	assert.False(t, ValidateFileType("text/plain"))
}

func TestValidateFileSize(t *testing.T) {
	err := ValidateFileSize("image/jpeg", 1024)
	assert.NoError(t, err)

	err = ValidateFileSize("image/jpeg", 6*1024*1024)
	assert.Error(t, err)

	err = ValidateFileSize("application/pdf", 1024)
	assert.NoError(t, err)

	err = ValidateFileSize("application/pdf", 21*1024*1024)
	assert.Error(t, err)
}

func TestBuildObjectKey(t *testing.T) {
	key := BuildObjectKey("contracts", 123, "contract.pdf")
	assert.Contains(t, key, "contracts")
	assert.Contains(t, key, "org_123")
	assert.Contains(t, key, "contract.pdf")
}

func TestNewClientWithInvalidCredentials(t *testing.T) {
	client, err := NewClient(Config{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		AccessKeyID:     "invalid",
		AccessKeySecret: "invalid",
		BucketName:      "nonexistent-bucket",
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
