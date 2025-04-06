package tokenizer

import (
	"encoding/hex"
	"errors"
	"crypto/rand"
	"time"

	"github.com/thelaumix/go-torken/serializer"
)

// Identifier for torkens
type Identifier struct {
    id []byte
}

func (id *Identifier) Bytes() []byte {
    return id.id
}
func (id *Identifier) Hex() string {
    if id.id == nil {
        return ""
    }
    return hex.EncodeToString(id.id)
}
func (id *Identifier) Anonymous() bool {
    return id.id == nil
}
func (id *Identifier) Valid() bool {
    return id.id != nil
}

// New random identifier
func NewIdentifier() *Identifier {
    token := make([]byte, 12)
    rand.Read(token)

    return &Identifier{
        id: token,
    }
}
// New anonymous identifier
func NewIdentifierAnonymous() *Identifier {
    return &Identifier{
        id: nil,
    }
}
// New identifier from byte array
func NewIdentifierFromBytes(id []byte) (*Identifier, error) {
    if id != nil && len(id) != cIDENTIFIER_LENGTH {
        return nil, errors.New("Identifier has to be either nil or 12 bytes long")
    }
    return &Identifier{
        id: id,
    }, nil
}
// New identifier from hex string
func NewIdentifierFromHex(hx string) (*Identifier, error) {
    if len(hx) == 0 {
        return &Identifier{
            id: nil,
        }, nil
    }

    data, err := hex.DecodeString(hx)
    if err != nil {return nil, err}
    if len(data) != cIDENTIFIER_LENGTH {
        return nil, errors.New("Identifier has to be either nil or 12 bytes long")
    }
    return &Identifier{
        id: data,
    }, nil
}


// Encrypt with an identifier
func (t *Tokenizer) Encrypt(payload any, key string, identifier *Identifier, options *encryptOptions) (string, error) {
    marshaled, err := serializer.Marshal(payload)
    if err != nil {return "", err}
    return t.int_encrypt(identifier.Bytes(), key, marshaled, options)
}

// Encrypt anonymous (without an identifier)
func (t *Tokenizer) EncryptAno(payload any, key string, options *encryptOptions) (string, error) {
    marshaled, err := serializer.Marshal(payload)
    if err != nil {return "", err}
    return t.int_encrypt(nil, key, marshaled, options)
}


type decryptHandle struct {
    err error
    data []byte
    i *decryptIntermediate
}

// Target container to decode the token payload into
func (h *decryptHandle) Into(outContainer any) (*tkResult, error) {
    if h.err != nil {
        return nil, h.err
    }
    err := serializer.Unmarshal(h.data, outContainer)
    if err != nil { return nil, err }

    var ID *Identifier
    if h.i.UsesIdentifier() {
        ID, err = NewIdentifierFromBytes(h.i.Identifier)
        if err != nil { return nil, err }
    } else {
        ID = NewIdentifierAnonymous()
    }

    return &tkResult{
        validFrom: time.Unix(int64(h.i.ValidFrom), 0),
        expiresIn: h.i.ExpiresIn,
        isValid: h.i.IsValid,
        identifier: *ID,
        version: uint8(h.i.Version()),
        algorithm: h.i.Algorithm(),
    }, nil
}

func decHandleErr(err error) *decryptHandle {
    return &decryptHandle{
        err: err,
        data: nil,
        i: nil,
    }
}

func decHandleBuf(i *decryptIntermediate, buf []byte) *decryptHandle {
    return &decryptHandle{
        err: nil,
        data: buf,
        i: i,
    }
}


type tkResult struct {
    validFrom   time.Time
    expiresIn   uint32
    isValid     bool
    identifier  Identifier
    version     uint8
    algorithm   TAlgorithm
}



// Time the token has been issued
func (r *tkResult) ValidFrom() time.Time { return r.validFrom }
// Seconds the token takes to expire. If `0`, it never does.
func (r *tkResult) ExpiresIn() uint32    { return r.expiresIn }
// Whether the token is valid
func (r *tkResult) IsValid() bool        { return r.isValid }
// The identifier for this token
func (r *tkResult) Identifier() Identifier { return r.identifier }
// Torken version this token has been created with
func (r *tkResult) Version() uint8       { return r.version }
// Used encryption algorithm
func (r *tkResult) Algorithm() TAlgorithm { return r.algorithm }


// Decrypt a buffer
func (t *Tokenizer) Decrypt(token, key string, options *decryptOptions) *decryptHandle {
    I, err := t.int_decrypt_begin(token, options)
    if err != nil {return decHandleErr(err)}
    byt, err := t.int_decrypt_finalize(I, key)
    if err != nil {return decHandleErr(err)}

    return decHandleBuf(I, byt)
}

// Decrypt a token with a keyResolver function.
func (t *Tokenizer) DecryptFn(token string, keyResolver func(identifier Identifier)(key string), options *decryptOptions) *decryptHandle {
    I, err := t.int_decrypt_begin(token, options)
    if err != nil {return decHandleErr(err)}
    var ID *Identifier
    var key string
    if I.UsesIdentifier() {
        ID, err = NewIdentifierFromBytes(I.Identifier)
        if err != nil {return decHandleErr(err)}
    } else {
        ID = NewIdentifierAnonymous()
    }
    key = keyResolver(*ID)
    byt, err := t.int_decrypt_finalize(I, key)
    if err != nil {return decHandleErr(err)}

    return decHandleBuf(I, byt)
}








