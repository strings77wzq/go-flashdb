package resp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
)

// Parser RESP协议解析器
type Parser struct {
	reader *bufio.Reader
}

// NewParser 创建新的解析器
func NewParser(reader io.Reader) *Parser {
	return &Parser{
		reader: bufio.NewReader(reader),
	}
}

// Parse 解析RESP协议
func (p *Parser) Parse() (Reply, error) {
	prefix, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch prefix {
	case SimpleStringPrefix:
		return p.parseSimpleString()
	case ErrorPrefix:
		return p.parseError()
	case IntegerPrefix:
		return p.parseInteger()
	case BulkPrefix:
		return p.parseBulk()
	case ArrayPrefix:
		return p.parseArray()
	default:
		return nil, errors.New("invalid RESP prefix")
	}
}

// parseSimpleString 解析简单字符串
func (p *Parser) parseSimpleString() (*SimpleStringReply, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}
	return &SimpleStringReply{Str: string(line)}, nil
}

// parseError 解析错误
func (p *Parser) parseError() (*ErrorReply, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}
	return &ErrorReply{Err: string(line)}, nil
}

// parseInteger 解析整数
func (p *Parser) parseInteger() (*IntegerReply, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}
	num, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return nil, err
	}
	return &IntegerReply{Num: num}, nil
}

// parseBulk 解析批量字符串
func (p *Parser) parseBulk() (*BulkReply, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}
	length, err := strconv.Atoi(string(line))
	if err != nil {
		return nil, err
	}
	if length == -1 {
		return NilBulkReply, nil
	}
	buf := make([]byte, length+2) // +2 for CRLF
	_, err = io.ReadFull(p.reader, buf)
	if err != nil {
		return nil, err
	}
	return &BulkReply{Arg: buf[:length]}, nil
}

// parseArray 解析数组
func (p *Parser) parseArray() (*ArrayReply, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}
	length, err := strconv.Atoi(string(line))
	if err != nil {
		return nil, err
	}
	if length == -1 {
		return NullArrayReply, nil
	}
	replies := make([]Reply, length)
	for i := 0; i < length; i++ {
		reply, err := p.Parse()
		if err != nil {
			return nil, err
		}
		replies[i] = reply
	}
	return &ArrayReply{Replies: replies}, nil
}

// readLine 读取一行（到CRLF结束）
func (p *Parser) readLine() ([]byte, error) {
	var line []byte
	for {
		b, err := p.reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		line = append(line, b...)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], nil // 去掉结尾的CRLF
}

// ParseCommand 解析Redis命令，返回命令名和参数
func ParseCommand(payload []byte) (string, [][]byte, error) {
	parser := NewParser(bytes.NewReader(payload))
	reply, err := parser.Parse()
	if err != nil {
		return "", nil, err
	}
	arrayReply, ok := reply.(*ArrayReply)
	if !ok {
		return "", nil, errors.New("command must be array")
	}
	if len(arrayReply.Replies) == 0 {
		return "", nil, errors.New("empty command")
	}
	cmdReply, ok := arrayReply.Replies[0].(*BulkReply)
	if !ok {
		return "", nil, errors.New("command name must be bulk string")
	}
	cmdName := string(bytes.ToLower(cmdReply.Arg))
	args := make([][]byte, 0, len(arrayReply.Replies)-1)
	for i := 1; i < len(arrayReply.Replies); i++ {
		argReply, ok := arrayReply.Replies[i].(*BulkReply)
		if !ok {
			return "", nil, errors.New("command argument must be bulk string")
		}
		args = append(args, argReply.Arg)
	}
	return cmdName, args, nil
}
