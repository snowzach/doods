package odrpc

import (
	"bytes"
	"encoding/base64"
)

// Images are byte arrays
type Raw []byte

// MarshalJSON for Raw fields is represented as base64
func (r *Raw) MarshalJSON() ([]byte, error) {
	if r == nil || *r == nil || len(*r) == 0 {
		return []byte(`""`), nil
	}
	return []byte(`"` + base64.StdEncoding.EncodeToString(*r) + `"`), nil
}

// UnmarshalJSON for Raw fields is parsed as base64
func (r *Raw) UnmarshalJSON(in []byte) error {
	var ret *[]byte
	if in == nil || len(in) == 0 {
		*ret = []byte{}
		return nil
	}
	// Remove the beginning and ending "
	in = bytes.Trim(in, `"`)
	*r = make([]byte, base64.StdEncoding.DecodedLen(len(in)))
	_, err := base64.StdEncoding.Decode(*r, in)
	return err
}
