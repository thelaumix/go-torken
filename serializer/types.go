package serializer

import (
	"golang.org/x/exp/constraints"
)

type serialType = uint8

const (
    ST_Unknown        serialType = 0x00

    ST_String         serialType = 0x01
    ST_Int8           serialType = 0x16
    ST_Uint8          serialType = 0x18
    ST_Int16          serialType = 0x17
    ST_Uint16         serialType = 0x19
    ST_Int32          serialType = 0x02
    ST_Uint32         serialType = 0x1a
    ST_Int64          serialType = 0x03
    ST_Uint64         serialType = 0x1b
    ST_Float16        serialType = 0x1c
    ST_Float          serialType = 0x04
    ST_Double         serialType = 0x05
    ST_Boolean        serialType = 0x06
    ST_Null           serialType = 0x07
    ST_Undefined      serialType = 0x08
    ST_Array          serialType = 0x09
    ST_Object         serialType = 0x0a
    ST_Bigint         serialType = 0x0b
    ST_Buffer         serialType = 0x0c
    ST_Function       serialType = 0x0d
    ST_Custom         serialType = 0x0e
    ST_Date           serialType = 0x0f

    // BUFFER WRAPPERS
    ST_Byte           serialType = 0x10
    ST_WORD           serialType = 0x11
    ST_DWORD          serialType = 0x12
    ST_LWORD          serialType = 0x13
    ST_MWORD          serialType = 0x14 // Custom shit (12 byte)
    ST_XWORD          serialType = 0x15 // Custom shit double LWord (16 byte)

    ST_Uuid           serialType = 0x1d
)

// Null-Equivalent type for decoding
type nullable interface{}
// Null-Equivalent type for decoding
var Null nullable

type Number interface {
    constraints.Integer | constraints.Float
}

type numberMode = byte
const (
    nm_Int    numberMode = 0
    nm_Uint   numberMode = 1
    nm_Float  numberMode = 2
)
