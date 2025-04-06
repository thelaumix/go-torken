package flexbuffer

import (
    "bytes"
)

type Flexbuf struct {
    buffer *bytes.Buffer
}

func (f *Flexbuf) Bytes() []byte {
    return f.buffer.Bytes()
}


func FromBytes(data *[]byte) *Flexbuf {
    return &Flexbuf{
        buffer: bytes.NewBuffer(*data),
    }
}


func Empty() *Flexbuf {
    return &Flexbuf{
        buffer: bytes.NewBuffer([]byte{}),
    }
}
