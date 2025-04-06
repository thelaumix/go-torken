package basex

import (
	"errors"
	"fmt"
)

// Encode some byte buffer with this transcoder
func (bx *BaseX) Encode(data []byte) (string, error) {
    if len(data) == 0 {
        return "", nil
    }

    zeroCount := 0
    idx := 0
    for idx < len(data) && data[idx] == 0 {
        idx++
        zeroCount++
    }

    // Ausgabegröße abschätzen
    size := int(float64(len(data)-idx)*bx.iFactor + 1)
    bX := make([]uint8, size)

    length := 0

    for idx < len(data) {
        carry := uint32(data[idx])
        i := 0
        // Von hinten durchmultiplizieren
        for it := size - 1; (carry != 0 || i < length) && it >= 0; it, i = it-1, i+1 {
            carry += 256 * uint32(bX[it])
            bX[it] = uint8(carry % uint32(bx.base))
            carry /= uint32(bx.base)
        }
        if carry != 0 {
            return "", errors.New("Non-zero carry while baseX encoding")
        }
        length = i
        idx++
    }

    // Leere Stellen überspringen
    it := size - length
    for it < size && bX[it] == 0 {
        it++
    }

    // Ergebnis zusammensetzen
    out := make([]byte, 0, zeroCount+(size-it))
    for i := 0; i < zeroCount; i++ {
        out = append(out, bx.leader)
    }
    for ; it < size; it++ {
        out = append(out, bx.alphabet[bX[it]])
    }
    return string(out), nil
}

// Decode some string with this transcoder
func (bx *BaseX) Decode(str string) ([]byte, error) {
    if len(str) == 0 {
        return nil, nil
    }

    zeroCount := 0
    idx := 0
    for idx < len(str) && str[idx] == bx.leader {
        zeroCount++
        idx++
    }

    size := int(float64(len(str)-idx)*bx.factor + 1)
    b256 := make([]uint8, size)

    length := 0
    for idx < len(str) {
        c := str[idx]
        carry := uint32(bx.baseMap[c])
        if carry == 255 {
            return nil, fmt.Errorf("invalid char %q for base %d", c, bx.base)
        }
        i := 0
        for it := size - 1; (carry != 0 || i < length) && it >= 0; it, i = it-1, i+1 {
            carry = carry + uint32(bx.base) * uint32(b256[it])
            b256[it] = uint8(carry % 256)
            carry = carry / 256

        }
        if carry != 0 {
            return nil, errors.New("non-zero carry")
        }
        length = i
        idx++
    }

    // Führende 0-Bytes überspringen
    it := size - length
    for it < size && b256[it] == 0 {
        it++
    }

    // Kopiere zeroCount + (size-it) Bytes in Ergebnis
    out := make([]byte, zeroCount+(size-it))
    copyIdx := 0
    // Erst Null-Bytes
    for i := 0; i < zeroCount; i++ {
        out[copyIdx] = 0
        copyIdx++
    }
    // Rest
    for it < size {
        out[copyIdx] = b256[it]
        copyIdx++
        it++
    }
    return out, nil
}
