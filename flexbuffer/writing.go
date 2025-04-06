package flexbuffer

import (
	"encoding/binary"
	"unsafe"
)

// Stored endianness to be shared for all instances
// TODO: Check if should really be little-endian
var NumericEndianness = binary.LittleEndian

/// Write a byte
func (f *Flexbuf) AppendByte(b byte) (error) {
    return f.buffer.WriteByte(b)
}

/// Write a numeric numeric value into flexbuffer
func AppendNumeric[T any](f *Flexbuf, val T) error {
    return binary.Write(f.buffer, NumericEndianness, val)
}

/// Append raw bytes to the buffer
func (f *Flexbuf) AppendRaw(ptr unsafe.Pointer, size int) (written int, err error) {
    slice := unsafe.Slice((*byte)(ptr), size)
    return f.buffer.Write(slice)
}

// Append a byte array to the buffer
func (f *Flexbuf) AppendBytes(data []byte) (written int, err error) {
    return f.buffer.Write(data)
}
