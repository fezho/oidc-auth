package server

import (
	"encoding/json"
	"fmt"
)

type (
	claims        map[string]json.RawMessage
	stringOrArray []string
)

func (c claims) unmarshalClaim(name string, v interface{}) error {
	val, ok := c[name]
	if !ok {
		return fmt.Errorf("claim not present")
	}
	return json.Unmarshal([]byte(val), v)
}

func (c claims) hasClaim(name string) bool {
	if _, ok := c[name]; !ok {
		return false
	}
	return true
}

func (c claims) extractUsername(usernameClaim string) (string, error) {
	var username string
	if err := c.unmarshalClaim(usernameClaim, &username); err != nil {
		return "", fmt.Errorf("oidc: parse username claims %q: %v", usernameClaim, err)
	}

	if usernameClaim == "email" {
		if err := c.verifyEmail(); err != nil {
			return "", err
		}
	}

	return username, nil
}

// verifyEmail ensures email is valid if the email_verified claim is present
// https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
func (c claims) verifyEmail() error {
	if hasEmailVerified := c.hasClaim("email_verified"); !hasEmailVerified {
		return nil
	}

	var emailVerified bool
	if err := c.unmarshalClaim("email_verified", &emailVerified); err != nil {
		return fmt.Errorf("oidc: parse 'email_verified' claim: %v", err)
	}

	// If the email_verified claim is present we have to verify it is set to `true`.
	if !emailVerified {
		return fmt.Errorf("oidc: email not verified")
	}

	return nil
}

func (c claims) extractGroups(groupsClaim string) ([]string, error) {
	if _, ok := c[groupsClaim]; ok {
		var groups stringOrArray
		if err := c.unmarshalClaim(groupsClaim, &groups); err != nil {
			return nil, fmt.Errorf("oidc: parse groups claim %q: %v", groupsClaim, err)
		}
		return []string(groups), nil
	}
	return nil, nil
}

func (s *stringOrArray) UnmarshalJSON(b []byte) error {
	var a []string
	if err := json.Unmarshal(b, &a); err == nil {
		*s = a
		return nil
	}
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	*s = []string{str}
	return nil
}
