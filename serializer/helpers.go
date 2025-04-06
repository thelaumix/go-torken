package serializer

import (
	"fmt"
	"math/big"
)

func er(format string, a ...any) (error) {
    return fmt.Errorf(format, a...)
}


func bigintToBuffer(bigint *big.Int) []byte {
    var (
        sign byte = 0
        bytes []byte
        dupe big.Int
    )

    dupe.Set(bigint)
    if bigint.Sign() < 0 {
        sign = 1
        dupe.Neg(bigint)
    }
    var FULL *big.Int = big.NewInt(0xff)

    bytes = append(bytes, sign)

    for dupe.Sign() > 0 {
        anded := big.NewInt(0)
        byt := byte(anded.And(&dupe, FULL).Int64())
        bytes = append(bytes, byt)
        dupe.Rsh(&dupe, 8)
    }

    return bytes
}

func bufferToBigint(buffer []byte) *big.Int {
    final := big.NewInt(0)
    sign := buffer[0]

    for i := (len(buffer)-1); i >= 1; i-- {
        final.Lsh(final, 8)
        final.Or(final, big.NewInt(int64(buffer[i])))
    }

    if sign == 1 {
        final.Neg(final)
    }

    return final
}
