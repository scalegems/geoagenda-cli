package store

import (
	"encoding/json"
	"errors"
	"fmt"

	keyring "github.com/zalando/go-keyring"

	"geoagenda-cli/internal/auth"
)

const (
	service = "geoagenda-cli"
	account = "default"
)

func Save(sess *auth.Session) error {
	b, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	return keyring.Set(service, account, string(b))
}

func Load() (*auth.Session, error) {
	raw, err := keyring.Get(service, account)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, fmt.Errorf("no stored session")
		}
		return nil, err
	}
	var s auth.Session
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil, fmt.Errorf("decode session: %w", err)
	}
	return &s, nil
}

func Clear() error {
	err := keyring.Delete(service, account)
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return err
	}
	return nil
}
