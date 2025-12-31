package security

import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha1"
    "encoding/base32"
    "encoding/binary"
    "fmt"
    "net/url"
    "strings"
    "time"
)

const defaultTOTPPeriod = 30
const totpDigits = 6

// GenerateTOTPSecret creates a new base32 encoded secret suitable for TOTP.
func GenerateTOTPSecret() (string, error) {
    buffer := make([]byte, 20)
    if _, err := rand.Read(buffer); err != nil {
        return "", err
    }
    // RFC 3548 padding removed for compatibility with authenticator apps.
    return strings.TrimRight(base32.StdEncoding.EncodeToString(buffer), "="), nil
}

// GenerateTOTPURI builds an otpauth URI for the provided secret and account.
func GenerateTOTPURI(secret, accountName, issuer string) string {
    issuer = strings.TrimSpace(issuer)
    if issuer == "" {
        issuer = "Atlas Wallet"
    }
    accountName = strings.TrimSpace(accountName)
    if accountName == "" {
        accountName = "user@atlaswallet"
    }

    label := url.QueryEscape(fmt.Sprintf("%s:%s", issuer, accountName))
    return fmt.Sprintf(
        "otpauth://totp/%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
        label,
        url.QueryEscape(secret),
        url.QueryEscape(issuer),
        totpDigits,
        defaultTOTPPeriod,
    )
}

// ValidateTOTP verifies a code against the secret with default tolerance.
func ValidateTOTP(secret, code string) bool {
    return ValidateTOTPWithWindow(secret, code, 1)
}

// ValidateTOTPWithWindow verifies a code allowing for configurable skew (number of periods).
func ValidateTOTPWithWindow(secret, code string, window int) bool {
    secret = strings.TrimSpace(secret)
    code = strings.TrimSpace(code)
    if secret == "" || code == "" {
        return false
    }

    if window < 0 {
        window = 0
    }

    key, err := decodeSecret(secret)
    if err != nil {
        return false
    }

    counter := time.Now().UTC().Unix() / defaultTOTPPeriod
    for offset := -window; offset <= window; offset++ {
        expected, err := generateCode(key, counter+int64(offset))
        if err != nil {
            continue
        }
        if subtleConstantTimeCompare(code, expected) {
            return true
        }
    }
    return false
}

func decodeSecret(secret string) ([]byte, error) {
    padding := len(secret) % 8
    if padding != 0 {
        secret += strings.Repeat("=", 8-padding)
    }
    return base32.StdEncoding.DecodeString(strings.ToUpper(secret))
}

func generateCode(secret []byte, counter int64) (string, error) {
    if counter < 0 {
        return "", fmt.Errorf("invalid counter")
    }
    counterBytes := make([]byte, 8)
    binary.BigEndian.PutUint64(counterBytes, uint64(counter))

    mac := hmac.New(sha1.New, secret)
    if _, err := mac.Write(counterBytes); err != nil {
        return "", err
    }
    sum := mac.Sum(nil)

    offset := sum[len(sum)-1] & 0x0f
    truncated := binary.BigEndian.Uint32(sum[offset : offset+4]) & 0x7fffffff
    code := truncated % 1000000

    return fmt.Sprintf("%06d", code), nil
}

func subtleConstantTimeCompare(a, b string) bool {
    if len(a) != len(b) {
        return false
    }
    result := byte(0)
    for i := 0; i < len(a); i++ {
        result |= a[i] ^ b[i]
    }
    return result == 0
}
