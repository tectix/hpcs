package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	SimpleString = '+'
	Error        = '-'
	Integer      = ':'
	BulkString   = '$'
	Array        = '*'
)

var (
	ErrInvalidProtocol = errors.New("invalid protocol")
	ErrInvalidType     = errors.New("invalid type")
)

type Value struct {
	Type  byte
	Str   string
	Int   int64
	Array []Value
}

func (v Value) Marshal() []byte {
	switch v.Type {
	case SimpleString:
		return []byte(fmt.Sprintf("+%s\r\n", v.Str))
	case Error:
		return []byte(fmt.Sprintf("-%s\r\n", v.Str))
	case Integer:
		return []byte(fmt.Sprintf(":%d\r\n", v.Int))
	case BulkString:
		if v.Str == "" {
			return []byte("$-1\r\n")
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v.Str), v.Str))
	case Array:
		if len(v.Array) == 0 {
			return []byte("*0\r\n")
		}
		result := []byte(fmt.Sprintf("*%d\r\n", len(v.Array)))
		for _, item := range v.Array {
			result = append(result, item.Marshal()...)
		}
		return result
	default:
		return []byte("-ERR unknown type\r\n")
	}
}

type Parser struct {
	reader *bufio.Reader
}

func NewParser(reader io.Reader) *Parser {
	return &Parser{
		reader: bufio.NewReader(reader),
	}
}

func (p *Parser) Parse() (Value, error) {
	typeByte, err := p.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch typeByte {
	case SimpleString:
		return p.parseSimpleString()
	case Error:
		return p.parseError()
	case Integer:
		return p.parseInteger()
	case BulkString:
		return p.parseBulkString()
	case Array:
		return p.parseArray()
	default:
		return Value{}, ErrInvalidType
	}
}

func (p *Parser) parseSimpleString() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: SimpleString, Str: line}, nil
}

func (p *Parser) parseError() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: Error, Str: line}, nil
}

func (p *Parser) parseInteger() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	
	num, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return Value{}, err
	}
	
	return Value{Type: Integer, Int: num}, nil
}

func (p *Parser) parseBulkString() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	
	length, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, err
	}
	
	if length == -1 {
		return Value{Type: BulkString, Str: ""}, nil
	}
	
	if length < 0 {
		return Value{}, ErrInvalidProtocol
	}
	
	data := make([]byte, length+2)
	_, err = io.ReadFull(p.reader, data)
	if err != nil {
		return Value{}, err
	}
	
	if data[length] != '\r' || data[length+1] != '\n' {
		return Value{}, ErrInvalidProtocol
	}
	
	return Value{Type: BulkString, Str: string(data[:length])}, nil
}

func (p *Parser) parseArray() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	
	count, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, err
	}
	
	if count == 0 {
		return Value{Type: Array, Array: []Value{}}, nil
	}
	
	if count < 0 {
		return Value{}, ErrInvalidProtocol
	}
	
	array := make([]Value, count)
	for i := 0; i < count; i++ {
		value, err := p.Parse()
		if err != nil {
			return Value{}, err
		}
		array[i] = value
	}
	
	return Value{Type: Array, Array: array}, nil
}

func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	
	if len(line) < 2 || line[len(line)-2] != '\r' || line[len(line)-1] != '\n' {
		return "", ErrInvalidProtocol
	}
	
	return strings.TrimSuffix(line, "\r\n"), nil
}

func NewSimpleString(s string) Value {
	return Value{Type: SimpleString, Str: s}
}

func NewError(s string) Value {
	return Value{Type: Error, Str: s}
}

func NewInteger(i int64) Value {
	return Value{Type: Integer, Int: i}
}

func NewBulkString(s string) Value {
	return Value{Type: BulkString, Str: s}
}

func NewArray(values ...Value) Value {
	return Value{Type: Array, Array: values}
}