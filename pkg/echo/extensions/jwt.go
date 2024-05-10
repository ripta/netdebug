package extensions

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/ripta/netdebug/pkg/echo/result"
)

var ErrMisconfigured = errors.New("misconfigured extension")

type JWTConfig struct {
	HeaderName        string
	JWKSURL           string
	IssuerURL         string
	Audience          string
	SigningAlgorithms []string
}

func JWT(jc JWTConfig) (result.ExtensionFunc, error) {
	if jc.HeaderName == "" || jc.JWKSURL == "" || jc.Audience == "" {
		return nil, ErrMisconfigured
	}

	ks := oidc.NewRemoteKeySet(context.Background(), jc.JWKSURL)
	cfg := oidc.Config{
		ClientID:             jc.Audience,
		SupportedSigningAlgs: jc.SigningAlgorithms,
	}

	verifier := oidc.NewVerifier(jc.IssuerURL, ks, &cfg)
	fn := func(r *http.Request) ([]result.ExtensionResult, error) {
		token := r.Header.Get(jc.HeaderName)
		if token == "" {
			return nil, nil
		}

		res := result.ExtensionResult{
			Name: "jwt:" + jc.HeaderName,
			Info: map[string][]string{},
		}

		idt, err := verifier.Verify(r.Context(), token)
		if err != nil {
			res.Info["error"] = []string{
				err.Error(),
			}
		} else {
			res.Info["subject"] = []string{idt.Subject}
			res.Info["issuer"] = []string{idt.Issuer}
			res.Info["audience"] = idt.Audience
			res.Info["issued_at"] = []string{idt.IssuedAt.Format(time.RFC3339Nano)}
			res.Info["expires_at"] = []string{idt.Expiry.Format(time.RFC3339Nano)}
			res.Info["nonce"] = []string{idt.Nonce}
			res.Info["access_token_hash"] = []string{idt.AccessTokenHash}
		}

		return []result.ExtensionResult{res}, nil
	}

	return fn, nil
}
