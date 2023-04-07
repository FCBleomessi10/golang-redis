package parser

import (
	"bufio"
	"errors"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

type Payload struct {
	Data resp.Reply
	Err  error
}

type readState struct {
	readingMultiLine  bool
	expectedArgsCount int
	msgType           byte
	args              [][]byte
	bulkLen           int64
}

func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

// ParseStream 异步地将解析结果送入管道, 不需要阻塞
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch) // 为每个用户生成一个解析器
	return ch
}

// parse 解析消息的主逻辑
func parse0(reader io.Reader, ch chan<- *Payload) {
	defer func() { // 不让for循环碰到异常退出
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()

	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	for true {
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			if ioErr { // IO错误
				ch <- &Payload{Err: err}
				close(ch)
				return
			}
			// 协议错误
			ch <- &Payload{Err: err}
			state = readState{}
			continue
		}

		if !state.readingMultiLine { // 判断是不是多行解析模式
			if msg[0] == '*' { // 数组
				err := parseMultiBulkHeader(msg, &state)
				if err != nil { // 协议错误
					ch <- &Payload{Err: errors.New("protocol error: " + string(msg))}
					state = readState{}
					continue
				}
				if state.expectedArgsCount == 0 { // 数组长度为0
					ch <- &Payload{Data: &reply.EmptyMultiBulkReply{}}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' { // 字符串
				err := parseBulkHeader(msg, &state)
				if err != nil { // 协议错误
					ch <- &Payload{Err: errors.New("protocol error: " + string(msg))}
					state = readState{}
					continue
				}
				if state.bulkLen == -1 { // 空字符串
					ch <- &Payload{Data: &reply.NullBulkReply{}}
					state = readState{}
					continue
				}
			} else { // 单行解析'+/-/:'
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{Data: result, Err: err}
				state = readState{}
				continue
			}
		} else {
			err := readBody(msg, &state)
			if err != nil {
				ch <- &Payload{Err: errors.New("protocol error: " + string(msg))}
				state = readState{}
				continue
			}

			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{Data: result, Err: err}
				state = readState{}
			}
		}
	}
}

func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error

	if state.bulkLen == 0 { // 1. \r\n切分
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}

		if len(msg) == 0 || msg[len(msg)-2] != '\r' { // 协议有问题
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else { // 2. 之前读到了数字, 严格读取字符个数
		msg = make([]byte, state.bulkLen+2)
		_, err := io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}

		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0
	}
	return msg, false, nil
}

func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32) // 解析命令的个数
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}

	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}

	if state.bulkLen == -1 {
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply

	switch msg[0] {
	case '+':
		result = reply.MakeStatusReply(str[1:])
	case '-':
		result = reply.MakeErrReply(str[1:])
	case ':':
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

func readBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2]

	var err error
	if line[0] == '$' {
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}

		if state.bulkLen <= 0 {
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
