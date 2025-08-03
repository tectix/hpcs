package protocol

import (
	"strconv"
	"strings"
	"time"

	"github.com/tectix/hpcs/internal/cache"
)

type CommandHandler struct {
	cache *cache.Cache
}

func NewCommandHandler(cache *cache.Cache) *CommandHandler {
	return &CommandHandler{
		cache: cache,
	}
}

func (h *CommandHandler) Execute(cmd Value) Value {
	if cmd.Type != Array || len(cmd.Array) == 0 {
		return NewError("ERR wrong number of arguments")
	}

	command := strings.ToUpper(cmd.Array[0].Str)
	args := cmd.Array[1:]

	switch command {
	case "GET":
		return h.handleGet(args)
	case "SET":
		return h.handleSet(args)
	case "DEL":
		return h.handleDel(args)
	case "EXISTS":
		return h.handleExists(args)
	case "KEYS":
		return h.handleKeys(args)
	case "FLUSHALL":
		return h.handleFlushAll(args)
	case "PING":
		return h.handlePing(args)
	case "INFO":
		return h.handleInfo(args)
	default:
		return NewError("ERR unknown command '" + command + "'")
	}
}

func (h *CommandHandler) handleGet(args []Value) Value {
	if len(args) != 1 {
		return NewError("ERR wrong number of arguments for 'get' command")
	}

	key := args[0].Str
	value, exists := h.cache.Get(key)
	if !exists {
		return Value{Type: BulkString, Str: ""}
	}

	return NewBulkString(string(value))
}

func (h *CommandHandler) handleSet(args []Value) Value {
	if len(args) < 2 {
		return NewError("ERR wrong number of arguments for 'set' command")
	}

	key := args[0].Str
	value := []byte(args[1].Str)
	var ttl time.Duration

	if len(args) > 2 {
		for i := 2; i < len(args); i++ {
			arg := strings.ToUpper(args[i].Str)
			switch arg {
			case "EX":
				if i+1 >= len(args) {
					return NewError("ERR syntax error")
				}
				seconds, err := strconv.Atoi(args[i+1].Str)
				if err != nil {
					return NewError("ERR value is not an integer or out of range")
				}
				ttl = time.Duration(seconds) * time.Second
				i++
			case "PX":
				if i+1 >= len(args) {
					return NewError("ERR syntax error")
				}
				ms, err := strconv.Atoi(args[i+1].Str)
				if err != nil {
					return NewError("ERR value is not an integer or out of range")
				}
				ttl = time.Duration(ms) * time.Millisecond
				i++
			}
		}
	}

	h.cache.Set(key, value, ttl)
	return NewSimpleString("OK")
}

func (h *CommandHandler) handleDel(args []Value) Value {
	if len(args) == 0 {
		return NewError("ERR wrong number of arguments for 'del' command")
	}

	deleted := int64(0)
	for _, arg := range args {
		if h.cache.Delete(arg.Str) {
			deleted++
		}
	}

	return NewInteger(deleted)
}

func (h *CommandHandler) handleExists(args []Value) Value {
	if len(args) == 0 {
		return NewError("ERR wrong number of arguments for 'exists' command")
	}

	exists := int64(0)
	for _, arg := range args {
		if _, found := h.cache.Get(arg.Str); found {
			exists++
		}
	}

	return NewInteger(exists)
}

func (h *CommandHandler) handleKeys(args []Value) Value {
	if len(args) != 1 {
		return NewError("ERR wrong number of arguments for 'keys' command")
	}

	pattern := args[0].Str
	keys := h.cache.Keys()

	var matchedKeys []Value
	for _, key := range keys {
		if matchPattern(key, pattern) {
			matchedKeys = append(matchedKeys, NewBulkString(key))
		}
	}

	return NewArray(matchedKeys...)
}

func (h *CommandHandler) handleFlushAll(args []Value) Value {
	h.cache.Clear()
	return NewSimpleString("OK")
}

func (h *CommandHandler) handlePing(args []Value) Value {
	if len(args) == 0 {
		return NewSimpleString("PONG")
	}
	return NewBulkString(args[0].Str)
}

func (h *CommandHandler) handleInfo(args []Value) Value {
	info := "# Server\r\n"
	info += "hpcs_version:1.0.0\r\n"
	info += "process_id:" + strconv.Itoa(1234) + "\r\n"
	info += "\r\n"
	info += "# Memory\r\n"
	info += "used_memory:" + strconv.FormatInt(h.cache.Size(), 10) + "\r\n"
	info += "keyspace_hits:0\r\n"
	info += "keyspace_misses:0\r\n"

	return NewBulkString(info)
}

func matchPattern(str, pattern string) bool {
	if pattern == "*" {
		return true
	}
	
	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
		return str == pattern
	}
	
	return simpleGlobMatch(str, pattern)
}

func simpleGlobMatch(str, pattern string) bool {
	if pattern == "" {
		return str == ""
	}
	
	if pattern == "*" {
		return true
	}
	
	if len(pattern) > 0 && pattern[0] == '*' {
		for i := 0; i <= len(str); i++ {
			if simpleGlobMatch(str[i:], pattern[1:]) {
				return true
			}
		}
		return false
	}
	
	if len(str) == 0 {
		return false
	}
	
	if len(pattern) > 0 && (pattern[0] == '?' || pattern[0] == str[0]) {
		return simpleGlobMatch(str[1:], pattern[1:])
	}
	
	return false
}