package serializer

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
	"time"

	"git.thelaumix.com/go-torken/flexbuffer"
	"github.com/google/uuid"
)




func setControlled(out *reflect.Value, use any, useWriting bool) error {
    if !useWriting {
        return nil
    }

    if !out.CanSet() {
        return er("Out target is not settable")
    }
    out.Set(reflect.ValueOf(use))
    return nil
}

func readNumeric[T Number](r *bytes.Reader) (t T, err error) {
    if err = binary.Read(r, flexbuffer.NumericEndianness, &t); err != nil {
        return
    }
    return
}

func readBytes(r *bytes.Reader, length uint32) (b []byte, err error) {
    b = make([]byte, length)
    _, err = io.ReadFull(r, b);
    return
}

func readBytesForLength(r *bytes.Reader) (b []byte, err error) {
    length, err := readNumeric[uint32](r)
    if err != nil {
        return
    }
    b, err = readBytes(r, length)
    return
}

func df_readArray(r *bytes.Reader, val *reflect.Value, typkind reflect.Kind, useWriting bool) error {
    // Anzahl der Elemente lesen
    count, err := readNumeric[uint32](r)
    if err != nil {
        return err
    }

    var (
        slice reflect.Value    = *val
        sliceType reflect.Type = val.Type().Elem()
    )

    if typkind == reflect.Array {
        if uint32(val.Len()) != count {
            return er("Inappropriate array size")
        }
    } else if typkind == reflect.Slice {
        // Neues Slice anlegen

        if uint32(val.Len()) != count {
            //slice.Grow(int(count))
            slice = reflect.MakeSlice(val.Type(), int(count), int(count))
        }
    } else if useWriting {
        return er("Inappropriate array container -> %s", typkind.String())
    }

    // Decode elements recursively
    for i := 0; i < int(count); i++ {
        if sliceType.Kind() == reflect.Interface || !useWriting {
            var anyVal interface{}
            if err := unmarshalFromReader(r, &anyVal, useWriting); err != nil {
                return err
            }
            if useWriting {
                slice.Index(i).Set(reflect.ValueOf(anyVal))
            }
        } else {
            elemPtr := slice.Index(i).Addr().Interface()
            if err := unmarshalFromReader(r, elemPtr, useWriting); err != nil {
                return err
            }
        }

    }
    val.Set(slice)

    return nil
}

func df_readObject(r *bytes.Reader, val *reflect.Value, typkind reflect.Kind, useWriting bool) error {
    // Anzahl Felder lesen
    var fieldCount uint32
    if err := binary.Read(r, binary.LittleEndian, &fieldCount); err != nil {
        return err
    }

    //var useInterface bool = false
    var targetMapVal *reflect.Value

    if typkind == reflect.Map {
        intermediateMap := reflect.MakeMap(val.Type())
        targetMapVal = &intermediateMap

    }

    // Für jedes Feld: Key lesen, Value entpacken
    for i := 0; i < int(fieldCount); i++ {
        // Key-Länge, Key-Bytes ...
        var keyLen uint32
        if err := binary.Read(r, binary.LittleEndian, &keyLen); err != nil {
            return err
        }
        keyBytes := make([]byte, keyLen)
        if _, err := io.ReadFull(r, keyBytes); err != nil {
            return err
        }
        key := string(keyBytes)

        if typkind == reflect.Map {
            // If is map, put in map at appropriate location
            elemPtr := reflect.New(val.Type().Elem())

            if err := unmarshalFromReader(r, elemPtr.Interface(), useWriting); err != nil {
                return err
            }
            targetMapVal.SetMapIndex(reflect.ValueOf(key), elemPtr.Elem())
        } else if typkind == reflect.Struct {
            // If is struct, put in struct
            fieldVal := findStructFieldByTorkenTag(*val, key)

            if !fieldVal.IsValid() {
                // ggf. Wert auslesen und verwerfen
                err := skipValue[map[string]any](r)
                if err != nil {
                    return err
                }
                continue
            }

            elemPtr := fieldVal.Addr().Interface()
            if err := unmarshalFromReader(r, elemPtr, useWriting); err != nil {
                return err
            }
        } else {
            return er("Inappropriate object container for %s -> %s", key, typkind.String())
        }
    }

    if useWriting && targetMapVal != nil {
        val.Set(*targetMapVal)
    }

    return nil
}



func df_Numeric[TNum Number](r *bytes.Reader, val *reflect.Value, mode numberMode, useWriting bool) error {
    num, err := readNumeric[TNum](r)
    if err != nil {
        return err
    }

    if val.Type().Kind() == reflect.Interface {
        return setControlled(val, num, useWriting)
    }
    if !useWriting {
        return nil
    }

    switch mode {
    case nm_Int:
        val.SetInt(int64(num))
    case nm_Uint:
        val.SetUint(uint64(num))
    case nm_Float:
        val.SetFloat(float64(num))
    default:
        return er("Failed to decompose numeric into target")
    }

    return nil
}

func df_Buffer(r *bytes.Reader, val *reflect.Value, len uint32, useWriting bool) error {
    b, err := readBytes(r, len)
    if err != nil {
        return nil
    }
    return setControlled(val, b, useWriting)
}



func unmarshalFromReader(r *bytes.Reader, out any, useWriting bool) error {

    // out muss Pointer auf eine Variable oder ein Struct sein
    val := reflect.ValueOf(out)
    if val.Kind() != reflect.Ptr || val.IsNil() {
        return er("invalid out pointer")
    }
    val = val.Elem()
    typ := val.Type()
    typkind := typ.Kind()

    // Fetch primary marshal type
    t, err := r.ReadByte()
    if err != nil {
        return err
    }

    // Switch decode branch
    switch t {
    case ST_String:
        b, err := readBytesForLength(r)
        if err != nil {
            return err
        }
        // Feld befüllen
        return setControlled(&val, string(b), useWriting)

    case ST_Uuid:
        b, err := readBytes(r, 16)
        if err != nil {
            return err
        }
        uid, err := uuid.FromBytes(b)
        if err != nil {
            return err
        }
        // Feld befüllen
        return setControlled(&val, uid, useWriting)

    case ST_Int8:       return df_Numeric[int8   ](r, &val, nm_Int  , useWriting)
    case ST_Uint8:      return df_Numeric[uint8  ](r, &val, nm_Uint , useWriting)
    case ST_Int16:      return df_Numeric[int16  ](r, &val, nm_Int  , useWriting)
    case ST_Uint16:     return df_Numeric[uint16 ](r, &val, nm_Uint , useWriting)
    case ST_Int32:      return df_Numeric[int32  ](r, &val, nm_Int  , useWriting)
    case ST_Uint32:     return df_Numeric[uint32 ](r, &val, nm_Uint , useWriting)
    case ST_Int64:      return df_Numeric[int64  ](r, &val, nm_Int  , useWriting)
    case ST_Uint64:     return df_Numeric[uint64 ](r, &val, nm_Uint , useWriting)
    case ST_Float:      return df_Numeric[float32](r, &val, nm_Float, useWriting)
    case ST_Double:     return df_Numeric[float64](r, &val, nm_Float, useWriting)

    case ST_Date:
        num, err := readNumeric[float64](r)
        if err != nil {
            return err
        }
        dat := time.UnixMilli(int64(num))
        return setControlled(&val, dat, useWriting)

    case ST_Buffer:
        blen, err := readNumeric[uint32](r)
        if err != nil {
            return err
        }
        return df_Buffer(r, &val, blen, useWriting)

    case ST_Byte:     return df_Buffer(r, &val, 1 , useWriting)
    case ST_WORD:     return df_Buffer(r, &val, 2 , useWriting)
    case ST_DWORD:    return df_Buffer(r, &val, 4 , useWriting)
    case ST_LWORD:    return df_Buffer(r, &val, 8 , useWriting)
    case ST_MWORD:    return df_Buffer(r, &val, 12, useWriting)
    case ST_XWORD:    return df_Buffer(r, &val, 16, useWriting)

    case ST_Boolean:
        b, err := r.ReadByte()
        if err != nil { return err }
        return setControlled(&val, b == 1, useWriting)

    case ST_Function:
        return skipValue[string](r)

    case ST_Unknown, ST_Undefined, ST_Null:
        return setControlled(&val, Null, useWriting)

    case ST_Array:
        return df_readArray(r, &val, typkind, useWriting)

    case ST_Object:
        return df_readObject(r, &val, typkind, useWriting)

    case ST_Bigint:
        buf, err := readBytesForLength(r)
        if err != nil { return err }

        bi := bufferToBigint(buf)

        return setControlled(&val, bi, useWriting)

    default:
        return er("Unexpected serial type")
    }
}

// findStructFieldByTorkenTag sucht im reflect.Value nach einem Feld mit Tag `torken:"key"`.
func findStructFieldByTorkenTag(structVal reflect.Value, key string) (reflect.Value) {
    typ := structVal.Type()
    for i := 0; i < typ.NumField(); i++ {
        sf := typ.Field(i)
        if sf.Tag.Get("torken") == key {
            return structVal.Field(i)
        }
    }
    return reflect.Value{}
}

// skipValue liest zur Not den aktuellen Value-Typ aus und verwirft entsprechende Daten.
// Damit kannst du unbekannte Felder auslassen.
func skipValue[TOut any](r *bytes.Reader) error {
    var dummyContainer TOut
    return unmarshalFromReader(r, &dummyContainer, false)
}


// Entmarschelt einen Maschelmarsch. Du Arsch!
//
// > "Wie barsch!"
func Unmarshal(encodedBuffer []byte, out any) error {
    r := bytes.NewReader(encodedBuffer)
    return unmarshalFromReader(r, out, true)
}
