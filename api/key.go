// Package api provides a standard key format for serialization to JSON or
// msgpack, and conversions to and from specific key types.
package api

import (
	"database/sql/driver"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Key is a concrete type for the keys.Key interface, which can be serialized
// and converted to concrete key types like keys.EdX25519Key.
// It also includes additional fields and metadata.
type Key struct {
	ID   keys.ID `json:"id,omitempty" msgpack:"id,omitempty" db:"id"`
	Type string  `json:"type,omitempty" msgpack:"type,omitempty" db:"type"`

	Private []byte `json:"priv,omitempty" msgpack:"priv,omitempty" db:"private"`
	Public  []byte `json:"pub,omitempty" msgpack:"pub,omitempty" db:"public"`

	CreatedAt int64 `json:"cts,omitempty" msgpack:"cts,omitempty" db:"createdAt"`
	UpdatedAt int64 `json:"uts,omitempty" msgpack:"uts,omitempty" db:"updatedAt"`

	// Optional fields
	Labels Labels `json:"labels,omitempty" msgpack:"labels,omitempty" db:"labels"`
	Notes  string `json:"notes,omitempty" msgpack:"notes,omitempty" db:"notes"`

	// Application specific fields
	Token   string `json:"token,omitempty" msgpack:"token,omitempty" db:"token"`
	Deleted bool   `json:"del,omitempty" msgpack:"del,omitempty" db:"del"`
}

// NewKey creates api.Key from keys.Key interface.
func NewKey(k keys.Key) *Key {
	return &Key{
		ID:      k.ID(),
		Public:  k.Public(),
		Private: k.Private(),
		Type:    string(k.Type()),
	}
}

// Created marks the key as created with the specified time.
func (k *Key) Created(ts int64) *Key {
	k.CreatedAt = ts
	k.UpdatedAt = ts
	return k
}

// Updated marks the key as created with the specified time.
func (k *Key) Updated(ts int64) *Key {
	k.UpdatedAt = ts
	return k
}

// WithLabels returns key with labels added.
func (k *Key) WithLabels(labels ...string) *Key {
	for _, label := range labels {
		if k.HasLabel(label) {
			return k
		}
		k.Labels = append(k.Labels, label)
	}
	return k
}

// HasLabel returns true if key has label.
func (k Key) HasLabel(label string) bool {
	for _, l := range k.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// WithNotes sets notes on key.
func (k *Key) WithNotes(notes string) *Key {
	k.Notes = notes
	return k
}

// Copy creates a copy of the key.
func (k *Key) Copy() *Key {
	b, err := msgpack.Marshal(k)
	if err != nil {
		return nil
	}
	var out Key
	if err := msgpack.Unmarshal(b, &out); err != nil {
		return nil
	}
	return &out
}

// Check if key is valid (has valid ID and type).
func (k *Key) Check() error {
	if k.ID == "" {
		return errors.Errorf("empty id")
	}
	if _, err := keys.ParseID(string(k.ID)); err != nil {
		return err
	}
	if k.Type == "" {
		return errors.Errorf("empty type")
	}
	if len(k.Public) == 0 && len(k.Private) == 0 {
		return errors.Errorf("no key data")
	}
	return nil
}

// Labels for key.
type Labels []string

// Scan for sql.DB.
func (p *Labels) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		spl := strings.Split(v, ",")
		out := []string{}
		for _, l := range spl {
			tr := strings.TrimSuffix(strings.TrimPrefix(l, "^"), "$")
			out = append(out, tr)
		}
		if len(out) != 0 {
			*p = out
		}
		return nil
	default:
		return errors.Errorf("invalid type for labels")
	}
}

// Value for sql.DB.
func (p Labels) Value() (driver.Value, error) {
	if len(p) == 0 {
		return driver.Value(""), nil
	}

	out := []string{}
	for _, l := range p {
		out = append(out, "^"+l+"$")
	}
	str := strings.Join(out, ",")
	return driver.Value(str), nil
}
