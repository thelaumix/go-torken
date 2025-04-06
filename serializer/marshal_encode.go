package serializer

import (
	"reflect"
	"time"

	"github.com/thelaumix/go-torken/flexbuffer"

	"math/big"

	"github.com/google/uuid"
)


func sf_String(data *flexbuffer.Flexbuf, t serialType, str string) (error) {
    err := data.AppendByte(t)
    if err != nil {
        return err
    }

    len := uint32(len(str))
    err = flexbuffer.AppendNumeric(data, len)
    if err != nil {
        return err
    }

    _, err = data.AppendBytes([]byte(str))
    if err != nil {
        return err
    }

    return nil
}

func sf_Numeric[T any](data *flexbuffer.Flexbuf, t serialType, num T) error {
    err := data.AppendByte(t)
    if err != nil {
        return err
    }

    err = flexbuffer.AppendNumeric(data, num)
    if err != nil {
        return err
    }

    return nil
}

func sf_ListBegin(data *flexbuffer.Flexbuf, t serialType, len uint32) error {
    err := data.AppendByte(t)
    if err != nil {
        return err
    }

    err = flexbuffer.AppendNumeric(data, len)
    return err
}

func sf_PushNamed(data *flexbuffer.Flexbuf, key string) (error) {

    len := uint32(len(key))
    err := flexbuffer.AppendNumeric(data, len)
    if err != nil {
        return err
    }

    _, err = data.AppendBytes([]byte(key))
    if err != nil {
        return err
    }

    return nil
}



func sf_Struct(data *flexbuffer.Flexbuf, val reflect.Value, typ reflect.Type) error {

    // Find fields with "torken" tag
    var torkenFields []reflect.StructField
    for i := 0; i < typ.NumField(); i++ {
        fieldTyp := typ.Field(i)
        // only use exported fields
        if fieldTyp.PkgPath != "" {
            continue
        }
        // Check for presence of "torken" tag
        if tagVal, ok := fieldTyp.Tag.Lookup("torken"); ok && tagVal != "" {
            torkenFields = append(torkenFields, fieldTyp)
        }
    }

    // Make object & len entry
    err := sf_ListBegin(data, ST_Object, uint32(len(torkenFields)))
    if err != nil {
        return err
    }

    // Now only iterate fields with a torken tag
    for _, fieldTyp := range torkenFields {
        fieldVal := val.FieldByName(fieldTyp.Name)
        // Tag-Name lesen
        torkenName := fieldTyp.Tag.Get("torken")

        // Use tag appended field name
        if err := sf_PushNamed(data, torkenName); err != nil {
            return err
        }

        // Serialize field value
        if err := serializeValue(data, fieldVal.Interface()); err != nil {
            return err
        }
    }

    return nil
}

func sf_Bigint(data *flexbuffer.Flexbuf, v *big.Int) error {
    buf := bigintToBuffer(v)
    err :=data.AppendByte(ST_Bigint)
    if err != nil { return err }
    err = flexbuffer.AppendNumeric(data, uint32(len(buf)))
    if err != nil { return err }
    _, err = data.AppendBytes(buf)
    if err != nil { return err }
    return nil
}



func serializeValue(data *flexbuffer.Flexbuf, v interface{}) (error) {
    val := reflect.ValueOf(v)
    typ := reflect.TypeOf(v)

    if typ.Kind() == reflect.Ptr {
        if val.IsNil() {
            return data.AppendByte(ST_Undefined)
        }
        val = val.Elem()
        typ = typ.Elem()
    }

	switch val.Kind() {
	case reflect.String:
        return sf_String(data, ST_String, val.String())
	// Weitere Cases für int, bool usw.
    case reflect.Int:        return sf_Numeric(data, ST_Int64 ,  int64( v.(int   )))
    case reflect.Uint:       return sf_Numeric(data, ST_Uint64,  uint64(v.(uint  )))
    case reflect.Int8:       return sf_Numeric(data, ST_Int8  ,  v.(int8  ))
    case reflect.Uint8:      return sf_Numeric(data, ST_Uint8 ,  v.(uint8 ))
    case reflect.Int16:      return sf_Numeric(data, ST_Int16 ,  v.(int16 ))
    case reflect.Uint16:     return sf_Numeric(data, ST_Uint16,  v.(uint16))
    case reflect.Int32:      return sf_Numeric(data, ST_Int32 ,  v.(int32 ))
    case reflect.Uint32:     return sf_Numeric(data, ST_Uint32,  v.(uint32))
    case reflect.Int64:      return sf_Numeric(data, ST_Int32 ,  v.(int64 ))
    case reflect.Uint64:     return sf_Numeric(data, ST_Uint32,  v.(uint64))
    case reflect.Float32:    return sf_Numeric(data, ST_Float ,  v.(float32))
    case reflect.Float64:    return sf_Numeric(data, ST_Double,  v.(float64))
    case reflect.Bool:
        boolSubstitute := uint8(0)
        if v.(bool) {
            boolSubstitute = uint8(1)
        }
        return sf_Numeric(data, ST_Boolean, boolSubstitute)

    case reflect.Array, reflect.Slice:
        len := uint32(val.Len())

        switch v := v.(type) {
        // UUID
        case uuid.UUID:
            err := data.AppendByte(ST_Uuid)
            if err != nil {
                return err
            }

            _, err = data.AppendBytes(v[0:16])
            return err

        // BYTE ARRAY
        case []byte, [1]byte, [2]byte, [4]byte, [8]byte, [12]byte, [16]byte: 
            fixlength := true
            var usingType serialType

            switch (len) {
            case 1: usingType  = ST_Byte
            case 2: usingType  = ST_WORD
            case 4: usingType  = ST_DWORD
            case 8: usingType  = ST_LWORD
            case 12: usingType = ST_MWORD
            case 16: usingType = ST_XWORD
            default:
                fixlength = false
                usingType = ST_Buffer
            }

            // Append type
            err := data.AppendByte(usingType)
            if err != nil {
                return err
            }

            if !fixlength {
                // Push length if non fix-length
                err = flexbuffer.AppendNumeric(data, len)
                if err != nil {
                    return err
                }
            }
            // Then iterate shit and process
            for _, byte := range val.Seq2() {
                err := data.AppendByte(byte.Interface().(uint8))
                if err != nil {
                    return err
                }
            }
            return nil

        // NORMAL ARRAY / SLICE
        default:
            // Write length first
            err := sf_ListBegin(data, ST_Array, len)
            if err != nil {
                return err
            }

            // Then iterate shit and process
            for _, el := range val.Seq2() {
                err := serializeValue(data, el.Interface())
                if err != nil {
                    return err
                }
            }
            return nil
        }


    case reflect.Map: 
        // Disallow non-string key maps
        if val.Type().Key().Kind() != reflect.String {
            return er("Cannot serialize maps with non-string key types")
        }

        // Write length first
        err := sf_ListBegin(data, ST_Object, uint32(val.Len()))
        if err != nil {
            return err
        }

        // Then iterate shit and process
        for key, el := range val.Seq2() {
            err := sf_PushNamed(data, key.String())
            if err != nil {
                return err
            }

            err = serializeValue(data, el.Interface())
            if err != nil {
                return err
            }
        }

        return nil


	case reflect.Struct:
        // TODO: Implement BigInt struct idea OR go internal big structure? Not sure yet
        switch v := v.(type) {
        case *big.Int:
            return sf_Bigint(data, v)
        case big.Int:
            return sf_Bigint(data, &v)
        case time.Time: 
            return sf_Numeric(data, ST_Date, float64(v.UnixMilli()))
        default:
            return sf_Struct(data, val, typ)
        }

    case reflect.Func:
        return er("Function serialization is intended for JS only and thus not supported here.");
	}

    switch v.(type) {
    case nullable: 
        return data.AppendByte(ST_Null)
    }

    return er("Unsupported type: %v", val.Kind())
}

func Marshal(v interface{}) ([]byte, error) {
    data := flexbuffer.Empty()
    err := serializeValue(data, v)
    if err != nil {
        return nil, err
    }

	return data.Bytes(), nil
}
