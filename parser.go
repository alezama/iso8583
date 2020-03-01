package iso8583

import (
	"errors"
	"fmt"
	"reflect"
)

// Parser for ISO 8583 messages
type Parser struct {
	messages     map[string]reflect.Type
	MtiEncode    int
	HeaderLength int
}

// Register MTI
func (p *Parser) Register(mti string, tpl interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Critical error:" + fmt.Sprint(r))
		}
	}()

	if len(mti) != 4 {
		return errors.New("MTI must be a 4 digit numeric field")
	}
	v := reflect.ValueOf(tpl)
	// TODO do more check
	if p.messages == nil {
		p.messages = make(map[string]reflect.Type)
	}
	p.messages[mti] = reflect.Indirect(v).Type()

	return nil
}

func decodeIsoHeader(raw []byte, isoHeaderLength int) (string, error) {
	if isoHeaderLength == 0 {
		return "", nil
	}
	var isoHeader string
	isoHeader = string(raw[:isoHeaderLength])
	return isoHeader, nil
}

func decodeMti(raw []byte, encode int, start int) (string, error) {
	mtiLen := 4
	if encode == BCD {
		mtiLen = 2
	}
	if len(raw) < mtiLen {
		return "", errors.New("bad MTI raw data")
	}

	var mti string
	switch encode {
	case ASCII:
		mti = string(raw[start : start+mtiLen])
	case BCD:
		mti = string(bcd2Ascii(raw[start : start+mtiLen]))
	default:
		return "", errors.New("invalid encode type")
	}
	return mti, nil
}

//Parse MTI
func (p *Parser) Parse(raw []byte) (ret *Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Critical error:" + fmt.Sprint(r))
			ret = nil
		}
	}()

	mti, err := decodeMti(raw, p.MtiEncode, p.HeaderLength)
	if err != nil {
		return nil, err
	}

	tp, ok := p.messages[mti]
	if !ok {
		return nil, errors.New("no template registered for MTI: " + mti)
	}
	tpl := reflect.New(tp)
	initStruct(tp, tpl)
	msg := NewMessage(mti, tpl.Interface())
	msg.MtiEncode = p.MtiEncode
	return msg, msg.Load(raw)
}

func initStruct(tp reflect.Type, val reflect.Value) {
	for i := 0; i < tp.NumField(); i++ {
		field := reflect.Indirect(val).Field(i)
		fieldType := tp.Field(i)
		switch fieldType.Type.Kind() {
		case reflect.Ptr: // only initialize Ptr fields
			fieldValue := reflect.New(fieldType.Type.Elem())
			field.Set(fieldValue)
		}
	}
}
