package tokenizer

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -ltorken_crypt_clink -lstdc++ -lssl -lcrypto -static -static-libstdc++ -static-libgcc
#include "crypt.h"
*/
import "C"
import (
	"crypto/sha256"
	"sort"
    "errors"
    "fmt"
    "unsafe"
)

// Dein Algorithm-Enum
type TAlgorithm byte

const (
    TALGO_CHACHA20 TAlgorithm = 0
    TALGO_AES      TAlgorithm = 1 // 256 GCM
)

// Binding to C-Version of the torken encryption algorithm
func crpEncrypt(algo TAlgorithm, in []byte, key string, nonce []byte) ([]byte, error) {
    // Wir nehmen an, dass out genauso groß ist wie in (ChaCha20 ohne GCM-Tag).
    // Bei AES-GCM kann der Ciphertext größer sein (GCM-Tag).
    // Passe ggf. die Größe an (len(in)+16 oder ähnliches).
    out := make([]byte, len(in))
    bytekey := []byte(key)

    if len(in) == 0 && len(key) == 0 && len(nonce) == 0 {
        return nil, errors.New("invalid parameters")
    }

    // Aufruf der C-Funktion
    rc := C.CRP_Encrypt_EX(
        C.int(algo),
        (*C.uint8_t)(unsafe.Pointer(&in[0])),    // in
        C.size_t(len(in)),
        (*C.uint8_t)(unsafe.Pointer(&bytekey[0])),   // key
        C.size_t(len(key)),
        (*C.uint8_t)(unsafe.Pointer(&nonce[0])), // nonce
        C.size_t(len(nonce)),
        (*C.uint8_t)(unsafe.Pointer(&out[0])),   // out
    )
    if rc != 0 {
        return nil, fmt.Errorf("CRP_Encrypt_EX failed with code %d", rc)
    }
    return out, nil
}

// Binding to C-Version of the torken decryption algorithm
func crpDecrypt(algo TAlgorithm, in []byte, key string, nonce []byte) ([]byte, error) {
    // Wir nehmen an, dass out genauso groß ist wie in (ChaCha20 ohne GCM-Tag).
    // Bei AES-GCM kann der Ciphertext größer sein (GCM-Tag).
    // Passe ggf. die Größe an (len(in)+16 oder ähnliches).
    out := make([]byte, len(in))
    bytekey := []byte(key)

    if len(in) == 0 && len(key) == 0 && len(nonce) == 0 {
        return nil, errors.New("invalid parameters")
    }

    // Aufruf der C-Funktion
    rc := C.CRP_Decrypt_EX(
        C.int(algo),
        (*C.uint8_t)(unsafe.Pointer(&in[0])),    // in
        C.size_t(len(in)),
        (*C.uint8_t)(unsafe.Pointer(&bytekey[0])),   // key
        C.size_t(len(key)),
        (*C.uint8_t)(unsafe.Pointer(&nonce[0])), // nonce
        C.size_t(len(nonce)),
        (*C.uint8_t)(unsafe.Pointer(&out[0])),   // out
    )
    if rc != 0 {
        return nil, fmt.Errorf("CRP_Encrypt_EX failed with code %d", rc)
    }
    return out, nil
}

// PseudoShuffle vertauscht data in-place basierend auf stable_sort nach key-Hash
func PseudoShuffle(data []byte, key string) {
    keyHash := sha256.Sum256([]byte(key))
    length := len(data)

    // Indizes vorbereiten
    indices := make([]int, length)
    for i := range indices {
        indices[i] = i
    }

    // stable_sort nach keyHash[i % 32]
    sort.SliceStable(indices, func(a, b int) bool {
        return keyHash[indices[a]%32] < keyHash[indices[b]%32]
    })

    // Reine Umordnung in temp
    temp := append([]byte(nil), data...)
    for i, idx := range indices {
        data[i] = temp[idx]
    }
}

// PseudoUnshuffle kehrt das Shuffle um
func PseudoUnshuffle(data []byte, key string) {
    keyHash := sha256.Sum256([]byte(key))
    length := len(data)

    // Indizes vorbereiten
    indices := make([]int, length)
    for i := range indices {
        indices[i] = i
    }
    sort.SliceStable(indices, func(a, b int) bool {
        return keyHash[indices[a]%32] < keyHash[indices[b]%32]
    })

    // Inversions-Tabelle
    reverseIndices := make([]int, length)
    for i, idx := range indices {
        reverseIndices[idx] = i
    }

    // Zurücksortieren
    temp := append([]byte(nil), data...)
    for i := 0; i < length; i++ {
        data[i] = temp[reverseIndices[i]]
    }
}

// makeChecksum: SHA256 über data und kopiere die ersten 8 Bytes in out (8 Byte)
func makeChecksum(out []byte, data []byte) {
    sum := sha256.Sum256(data)
    copy(out, sum[:8])
}
