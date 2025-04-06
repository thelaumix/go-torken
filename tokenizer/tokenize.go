package tokenizer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"git.thelaumix.com/go-torken/basex"
)

// Nur als Beispiel: Du könntest IDENTIFIER_LENGTH = 12 übernehmen.
const cIDENTIFIER_LENGTH = 12

// TOKENIZER_LATEST_VERSION analog zum #define
const cTOKENIZER_LATEST_VERSION = 0x01

// encryptOptions dient als Pendant zu C++-encryptOptions. Alles optional per *.
type encryptOptions struct {
    validFrom  *uint32
    expiresIn  *uint32
    scrambler  *string
    algorithm  TAlgorithm
    version    uint8
    alphabet   *string
}
func (o *encryptOptions) ValidFrom(v time.Time) { 
    ts := uint32(v.Unix())
    o.validFrom = &ts 
}
func (o *encryptOptions) ValidFromTS(v uint32)   *encryptOptions { o.validFrom = &v; return o }
func (o *encryptOptions) ExpiresIn(v uint32)     *encryptOptions { o.validFrom = &v; return o }
func (o *encryptOptions) Scrambler(v string)     *encryptOptions { o.scrambler = &v; return o }
func (o *encryptOptions) Algorithm(v TAlgorithm) *encryptOptions { o.algorithm = v; return o }
func (o *encryptOptions) ChaCha20()              *encryptOptions { o.algorithm = TALGO_CHACHA20; return o }
func (o *encryptOptions) AES()                   *encryptOptions { o.algorithm = TALGO_AES; return o }
func (o *encryptOptions) Version(v uint8)        *encryptOptions { o.version = v;   return o }
func (o *encryptOptions) Alphabet(v string)      *encryptOptions { o.alphabet = &v;  return o }

func EncryptOptions() *encryptOptions {
    return &encryptOptions{
        version: cTOKENIZER_LATEST_VERSION,
        algorithm: TALGO_CHACHA20,
    }
}

// decryptOptions analog
type decryptOptions struct {
    scrambler *string
    alphabet  *string
}
func (o *decryptOptions) Scrambler(v string)     *decryptOptions { o.scrambler = &v; return o }
func (o *decryptOptions) Alphabet(v string)      *decryptOptions { o.alphabet = &v;  return o }

func DecryptOptions() *decryptOptions {
    return &decryptOptions{}
}

// decryptIntermediate entspricht dem struct in C++
type decryptIntermediate struct {
    Vhead           byte
    Identifier      []byte
    Nonce           []byte
    EncryptedPayload []byte
    ValidFrom       uint32
    ExpiresIn       uint32
    IsValid         bool
    // hier könnte man wie im C++-Code I.type usw. abbilden
    PayloadType     byte
}

// Methode zum Herauslesen, ob Identifiert benutzt wird
func (d *decryptIntermediate) UsesIdentifier() bool {
    // vhead 7. Bit
    return (d.Vhead>>7)&0x01 == 1
}

func (d *decryptIntermediate) Version() uint8 {
    return uint8(d.Vhead & 0b1111)
}

// Algorithm-Getter (entspricht I.algorithm() in C++)
func (d *decryptIntermediate) Algorithm() TAlgorithm {
    // Bits 4..6 in vhead? Hier vereinfacht: shift 4
    return TAlgorithm((d.Vhead >> 4) & 0x07)
}

// Tokenizer entspricht der C++-Klasse
type Tokenizer struct {
    scramblerKey string
    baseX        *basex.BaseX // das BaseX-Gerüst aus dem vorherigen Beispiel
}

// NewTokenizer als Konstruktor-Ersatz
func NewTokenizer() *Tokenizer {
    return &Tokenizer{
        scramblerKey: "DEFAULT_SCRAMBLER", // entspricht DEFAULT_SCRAMBLER
        baseX:        basex.NewBaseXDefault(),  // dein BaseX-Default-Konstruktor
    }
}

// SetAlphabet analog zur C++-Methode
func (t *Tokenizer) SetAlphabet(alphabet string) error {
    if len(alphabet) < 2 {
        return errors.New("invalid alphabet length")
    }
    return t.baseX.SetAlphabet(alphabet)
}

// SetScrambler analog
func (t *Tokenizer) SetScrambler(scrambler string) {
    t.scramblerKey = scrambler
}

// GetScrambler analog
func (t *Tokenizer) GetScrambler() string {
    return t.scramblerKey
}

// GetAlphabet analog
func (t *Tokenizer) GetAlphabet() string {
    return t.baseX.GetAlphabet()
}

// int_encrypt grob nach dem Vorbild des C++-Codes.
// – identifier (ggf. nil)
// – key
// – payload
// – options
// – outResult
func (t *Tokenizer) int_encrypt(identifier []byte, key string, payload []byte,
    options *encryptOptions) (string, error) {

    data := append([]byte(nil), payload...) // Payload kopieren

    validFrom := uint32(time.Now().Unix())
    expiresIn := uint32(0)
    algorithm := TALGO_CHACHA20
    version := uint8(cTOKENIZER_LATEST_VERSION)
    useIdentifier := identifier != nil
    scrambkey := t.scramblerKey

    // Falls options != nil, Felder ggf. überschreiben
    if options != nil {
        algorithm = options.algorithm
        if options.validFrom != nil {
            validFrom = *options.validFrom
        }
        if options.expiresIn != nil {
            expiresIn = *options.expiresIn
        }
        if options.scrambler != nil {
            scrambkey = *options.scrambler
        }
        version = options.version
        if options.alphabet != nil {
            _ = t.baseX.SetAlphabet(*options.alphabet)
        }
    }

    // Vhead zusammenbauen
    // vhead = version + (algorithm << 4) + (useIdentifier << 7)
    vhead := version
    vhead |= byte(algorithm << 4)
    if useIdentifier {
        vhead |= 1 << 7
    }

    // Checksum Input (vhead + identifier + validFrom + expiresIn + payload)
    var checksumInput bytes.Buffer
    _ = checksumInput.WriteByte(vhead)
    if useIdentifier {
        checksumInput.Write(identifier)
    }
    binary.Write(&checksumInput, binary.LittleEndian, validFrom)
    binary.Write(&checksumInput, binary.LittleEndian, expiresIn)
    checksumInput.Write(data)

    checksum := make([]byte, 8)
    makeChecksum(checksum, checksumInput.Bytes()) // -> externes Hilfsfunktion

    // Nonce = validFrom + expiresIn + checksum (jeweils 4,4,8)
    nonce := make([]byte, 16)
    binary.LittleEndian.PutUint32(nonce[0:], validFrom)
    binary.LittleEndian.PutUint32(nonce[4:], expiresIn)
    copy(nonce[8:], checksum)

    // Verschlüsseln
    encrypted, err := crpEncrypt(algorithm, data, key, nonce)
    //err := crpEncrypt(algorithm, encrypted, data, key, nonce)
    if err != nil {
        return "", err
    }

    // finalBuffer = [vhead][identifier][nonce][encrypted]
    var finalBuffer bytes.Buffer
    finalBuffer.WriteByte(vhead)
    if useIdentifier {
        finalBuffer.Write(identifier)
    }
    finalBuffer.Write(nonce)
    finalBuffer.Write(encrypted)

    // PseudoShuffle
    PseudoShuffle(finalBuffer.Bytes(), scrambkey) // Stub-Funktion

    // BaseX-encode
    outResult, err := t.baseX.Encode(finalBuffer.Bytes())
    if err != nil {
        return "", err
    }
    return outResult, nil
}

// int_decrypt_begin entspricht Decrypt_Begin
func (t *Tokenizer) int_decrypt_begin(tokenString string, options *decryptOptions) (*decryptIntermediate, error) {
    scrambkey := t.scramblerKey

    if options != nil {
        if options.scrambler != nil {
            scrambkey = *options.scrambler
        }
        if options.alphabet != nil {
            _ = t.baseX.SetAlphabet(*options.alphabet)
        }
    }

    encryptedData, err := t.baseX.Decode(tokenString)
    if err != nil {
        return nil, err
    }

    // Unshuffle
    PseudoUnshuffle(encryptedData, scrambkey)

    if len(encryptedData) < 1 {
        return nil, errors.New("corrupt data: too short")
    }

    I := &decryptIntermediate{}
    I.Vhead = encryptedData[0]

    usesIdentifier := I.UsesIdentifier()
    expectedIDSize := 0
    if usesIdentifier {
        expectedIDSize = cIDENTIFIER_LENGTH
    }

    minLen := 1 + expectedIDSize + 16
    if len(encryptedData) < minLen {
        return nil, errors.New("corrupt data: length too small")
    }

    // identifier
    I.Identifier = append([]byte(nil), encryptedData[1:1+expectedIDSize]...)
    // nonce
    nonceStart := 1 + expectedIDSize
    I.Nonce = append([]byte(nil), encryptedData[nonceStart:nonceStart+16]...)
    // rest = encryptedPayload
    I.EncryptedPayload = append([]byte(nil), encryptedData[nonceStart+16:]...)

    I.ValidFrom = binary.LittleEndian.Uint32(I.Nonce[0:4])
    I.ExpiresIn = binary.LittleEndian.Uint32(I.Nonce[4:8])

    now := uint32(time.Now().Unix())
    // expiresIn == 0 => kein Ablauf
    I.IsValid = (I.ExpiresIn == 0 || (now >= I.ValidFrom && now < I.ValidFrom+I.ExpiresIn))

    return I, nil
}

// int_decrypt_finalize entspricht Decrypt_Finalize
// Man übergibt das Intermediate und bekommt die entschlüsselte Payload zurück
func (t *Tokenizer) int_decrypt_finalize(I *decryptIntermediate, key string) ([]byte, error) {

    decrypted, err := crpDecrypt(I.Algorithm(), I.EncryptedPayload, key, I.Nonce)
    if err != nil {
        return nil, err
    }

    // Checksum prüfen
    var checkBuf bytes.Buffer
    checkBuf.WriteByte(I.Vhead)
    if I.UsesIdentifier() {
        checkBuf.Write(I.Identifier)
    }
    binary.Write(&checkBuf, binary.LittleEndian, I.ValidFrom)
    binary.Write(&checkBuf, binary.LittleEndian, I.ExpiresIn)
    checkBuf.Write(decrypted)

    computed := make([]byte, 8)
    makeChecksum(computed, checkBuf.Bytes())

    // I.Nonce[8..16] == checksum
    if !bytes.Equal(computed, I.Nonce[8:16]) {
        return nil, errors.New("token integrity invalid")
    }

    // „Type“ = decryptedPayload[0], analog I.type = ...
    if len(decrypted) > 0 {
        I.PayloadType = decrypted[0]
    }
    // z.B. den Rest als eigentliche Payload
    return decrypted, nil
}

