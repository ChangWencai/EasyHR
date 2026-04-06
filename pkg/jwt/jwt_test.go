package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-for-jwt-testing-only-must-be-long-enough"

func TestGenerateAndParseAccessToken(t *testing.T) {
	token, err := GenerateAccessToken(1, 100, "owner", testSecret, time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := ParseToken(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
	assert.Equal(t, int64(100), claims.OrgID)
	assert.Equal(t, "owner", claims.Role)
	assert.NotEmpty(t, claims.ID)
}

func TestParseExpiredToken(t *testing.T) {
	token, _ := GenerateAccessToken(1, 100, "owner", testSecret, -time.Hour)
	_, err := ParseToken(token, testSecret)
	assert.Error(t, err)
}

func TestParseTamperedToken(t *testing.T) {
	token, _ := GenerateAccessToken(1, 100, "owner", testSecret, time.Hour)
	tampered := token[:len(token)-5] + "xxxxx"
	_, err := ParseToken(tampered, testSecret)
	assert.Error(t, err)
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken(1, testSecret, time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := ParseRefreshToken(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
	assert.NotEmpty(t, claims.ID)
}

func TestRefreshTokens(t *testing.T) {
	refreshToken, _ := GenerateRefreshToken(1, testSecret, time.Hour)

	var blacklistedJTI string
	blacklistFunc := func(jti string, ttl time.Duration) error {
		blacklistedJTI = jti
		return nil
	}

	pair, err := RefreshTokens(refreshToken, testSecret, time.Minute, time.Hour, blacklistFunc)
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEqual(t, refreshToken, pair.RefreshToken)
	assert.NotEmpty(t, blacklistedJTI)
}

func TestRefreshTokensWithInvalidToken(t *testing.T) {
	blacklistFunc := func(jti string, ttl time.Duration) error { return nil }
	_, err := RefreshTokens("invalid.token.here", testSecret, time.Minute, time.Hour, blacklistFunc)
	assert.Error(t, err)
}

func TestRefreshTokensBlacklistError(t *testing.T) {
	refreshToken, _ := GenerateRefreshToken(1, testSecret, time.Hour)
	blacklistFunc := func(jti string, ttl time.Duration) error {
		return assert.AnError
	}
	_, err := RefreshTokens(refreshToken, testSecret, time.Minute, time.Hour, blacklistFunc)
	assert.Error(t, err)
}
