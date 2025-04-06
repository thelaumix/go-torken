package basex

import (
    "errors"
)

const (
    defaultAlphabet string      = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

// BaseX entspricht grob deiner C++-Klasse.
type BaseX struct {
    alphabet string
    baseMap  [256]uint8
    base     int
    leader   byte
    factor   float64 // log(Base) / log(256)
    iFactor  float64 // log(256) / log(Base)
}

// Construct a new BaseX transcoder with a custom alphabet
func NewBaseX(alphabet string) (*BaseX, error) {
    if len(alphabet) >= 255 {
        return nil, errors.New("alphabet too long")
    }
    bx := &BaseX{}
    if err := bx.SetAlphabet(alphabet); err != nil {
        return nil, err
    }
    return bx, nil
}

// Construct a new BaseX transcoder with the default alphabet
func NewBaseXDefault() *BaseX {
    bx := &BaseX{}
    _ = bx.SetAlphabet(defaultAlphabet)
    return bx
}
