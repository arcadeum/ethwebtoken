package ethwebtoken

import (
	"fmt"
	"time"

	"github.com/arcadeum/ethkit/ethcoder"
)

type Token struct {
	// "eth" prefix
	Prefix string

	// Account addres (in hex)
	Address string

	// Claims object, aka, the message key of an EIP712 signature
	Claims Claims

	// Signature of the message by the account address above (in hex)
	Signature string
}

func NewToken() *Token {
	return &Token{
		Prefix: EWTPrefix,
		Claims: Claims{
			EWTVersion: EWTVersion,
		},
	}
}

func (t *Token) MessageDigest() ([]byte, error) {
	return t.Claims.MessageDigest()
}

func (t *Token) MessageTypedData() (*ethcoder.TypedData, error) {
	return t.Claims.TypedData()
}

type Claims struct {
	App        string `json:"app,omitempty"`
	IssuedAt   int64  `json:"iat,omitempty"`
	ExpiresAt  int64  `json:"exp,omitempty"`
	Nonce      uint64 `json:"n,omitempty"`
	Type       string `json:"typ,omitempty"`
	Origin     string `json:"ogn,omitempty"`
	EWTVersion string `json:"v,omitempty"`
}

func (c *Claims) SetIssuedAtNow() {
	c.IssuedAt = time.Now().UTC().Unix()
}

func (c *Claims) SetExpiryIn(tm time.Duration) {
	c.ExpiresAt = time.Now().UTC().Unix() + int64(tm.Seconds())
}

func (c Claims) Valid() error {
	now := time.Now().Unix()
	drift := int64(5 * 60)                                                // 5 minutes
	max := int64(time.Duration((time.Hour * 24 * 365) + drift).Seconds()) // 1 year

	if c.EWTVersion == "" {
		return fmt.Errorf("claims: ewt version is empty")
	}
	if c.App == "" {
		return fmt.Errorf("claims: app is empty")
	}
	if c.IssuedAt > now+drift || c.IssuedAt < now-max {
		return fmt.Errorf("claims: iat is invalid")
	}
	if c.ExpiresAt < now-drift || c.ExpiresAt > now+max {
		return fmt.Errorf("claims: token has expired")
	}

	return nil
}

func (c Claims) Map() map[string]interface{} {
	m := map[string]interface{}{}
	if c.App != "" {
		m["app"] = c.App
	}
	if c.IssuedAt != 0 {
		m["iat"] = c.IssuedAt
	}
	if c.ExpiresAt != 0 {
		m["exp"] = c.ExpiresAt
	}
	if c.Nonce != 0 {
		m["n"] = c.Nonce
	}
	if c.Type != "" {
		m["typ"] = c.Type
	}
	if c.Origin != "" {
		m["ogn"] = c.Origin
	}
	if c.EWTVersion != "" {
		m["v"] = c.EWTVersion
	}
	return m
}

func (c Claims) TypedData() (*ethcoder.TypedData, error) {
	td := &ethcoder.TypedData{
		Types: ethcoder.TypedDataTypes{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
			},
			"Claims": {},
		},
		PrimaryType: "Claims",
		Domain:      eip712Domain,
		Message:     c.Map(),
	}

	if len(td.Message) == 0 {
		return nil, fmt.Errorf("ethwebtoken: claims is empty")
	}

	claimsType := []ethcoder.TypedDataArgument{}
	if c.App != "" {
		claimsType = append(claimsType, ethcoder.TypedDataArgument{Name: "app", Type: "string"})
	}
	if c.IssuedAt != 0 {
		claimsType = append(claimsType, ethcoder.TypedDataArgument{Name: "iat", Type: "int64"})
	}
	if c.ExpiresAt != 0 {
		claimsType = append(claimsType, ethcoder.TypedDataArgument{Name: "exp", Type: "int64"})
	}
	if c.Nonce != 0 {
		claimsType = append(claimsType, ethcoder.TypedDataArgument{Name: "n", Type: "uint64"})
	}
	if c.Type != "" {
		claimsType = append(claimsType, ethcoder.TypedDataArgument{Name: "typ", Type: "string"})
	}
	if c.Origin != "" {
		claimsType = append(claimsType, ethcoder.TypedDataArgument{Name: "ogn", Type: "string"})
	}
	if c.EWTVersion != "" {
		claimsType = append(claimsType, ethcoder.TypedDataArgument{Name: "v", Type: "string"})
	}
	td.Types["Claims"] = claimsType

	return td, nil
}

func (c Claims) MessageDigest() ([]byte, error) {
	if err := c.Valid(); err != nil {
		return nil, fmt.Errorf("claims are invalid - %w", err)
	}

	typedData, err := c.TypedData()
	if err != nil {
		return nil, fmt.Errorf("ethwebtoken: failed to compute claims typed data - %w", err)
	}
	digest, err := typedData.EncodeDigest()
	if err != nil {
		return nil, fmt.Errorf("ethwebtoken: failed to compute claims message digest - %w", err)
	}
	return digest, nil
}
