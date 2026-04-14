package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	statusAccessVersion     = "v2"
	statusAccessTokenExpiry = 24 * time.Hour
)

// CreateOrderStatusAccessToken generates a time-limited, HMAC-signed token
// that allows anonymous access to a specific order's status page.
// Format: "{expiresAt}.{userId}.{signature}"
// Matches the TypeScript status-access.ts implementation.
func CreateOrderStatusAccessToken(orderID, userID int64, adminToken string) string {
	expiresAt := time.Now().Add(statusAccessTokenExpiry).Unix()
	sig := computeStatusAccessSig(orderID, userID, expiresAt, adminToken)
	return fmt.Sprintf("%d.%d.%s", expiresAt, userID, sig)
}

// VerifyOrderStatusAccessToken validates a status access token for a given order.
// Returns the user ID from the token if valid, or an error if invalid/expired.
func VerifyOrderStatusAccessToken(orderID int64, token, adminToken string) (int64, error) {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid token format")
	}

	expiresAt, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid expiration")
	}
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID")
	}
	sig := parts[2]

	// Check expiration
	if time.Now().Unix() > expiresAt {
		return 0, fmt.Errorf("token expired")
	}

	// Verify signature with constant-time comparison
	expected := computeStatusAccessSig(orderID, userID, expiresAt, adminToken)
	if subtle.ConstantTimeCompare([]byte(sig), []byte(expected)) != 1 {
		return 0, fmt.Errorf("invalid signature")
	}

	return userID, nil
}

// computeStatusAccessSig computes the HMAC-SHA256 signature for a status access token.
func computeStatusAccessSig(orderID, userID, expiresAt int64, adminToken string) string {
	// Derive key: HMAC-SHA256("order-status-access-key", adminToken)
	keyMac := hmac.New(sha256.New, []byte(adminToken))
	keyMac.Write([]byte("order-status-access-key"))
	derivedKey := keyMac.Sum(nil)

	// Sign: HMAC-SHA256(message, derivedKey)
	message := fmt.Sprintf("order-status-access:%s:%d:%d:%d", statusAccessVersion, orderID, userID, expiresAt)
	mac := hmac.New(sha256.New, derivedKey)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
