package basex

import (
	"errors"
	"fmt"
	"math"
)

// SetAlphabet
func (bx *BaseX) SetAlphabet(alphabet string) error {
    if len(alphabet) < 2 {
        return errors.New("Alphabet must be >= 2 chars")
    }
    bx.alphabet = alphabet
    bx.base = len(alphabet)
    bx.leader = alphabet[0]

    // BaseMap initialisieren
    for i := range bx.baseMap {
        bx.baseMap[i] = 255
    }
    for i := 0; i < bx.base; i++ {
        c := alphabet[i]
        if bx.baseMap[c] != 255 {
            return fmt.Errorf("%q is ambiguous in alphabet", c)
        }
        bx.baseMap[c] = uint8(i)
    }

    // Log-Faktoren
    bx.factor = math.Log(float64(bx.base)) / math.Log(256)
    bx.iFactor = math.Log(256) / math.Log(float64(bx.base))
    return nil
}

// GetAlphabet
func (bx *BaseX) GetAlphabet() string {
    return bx.alphabet
}

// GetBase
func (bx *BaseX) GetBase() int {
    return bx.base
}
