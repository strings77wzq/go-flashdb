package core

import (
	"sync"

	"goflashdb/pkg/resp"
)

var (
	pubSubManager *PubSubManager
)

type PubSubManager struct {
	mu       sync.RWMutex
	channels map[string]map[*Subscriber]bool
	patterns map[string]map[*Subscriber]bool
}

type Subscriber struct {
	ch     chan []byte
	closed bool
}

func NewPubSubManager() *PubSubManager {
	return &PubSubManager{
		channels: make(map[string]map[*Subscriber]bool),
		patterns: make(map[string]map[*Subscriber]bool),
	}
}

func GetPubSubManager() *PubSubManager {
	if pubSubManager == nil {
		pubSubManager = NewPubSubManager()
	}
	return pubSubManager
}

func (pm *PubSubManager) Subscribe(channel string, sub *Subscriber) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.channels[channel] == nil {
		pm.channels[channel] = make(map[*Subscriber]bool)
	}
	pm.channels[channel][sub] = true
}

func (pm *PubSubManager) PSubscribe(pattern string, sub *Subscriber) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.patterns[pattern] == nil {
		pm.patterns[pattern] = make(map[*Subscriber]bool)
	}
	pm.patterns[pattern][sub] = true
}

func (pm *PubSubManager) Unsubscribe(channel string, sub *Subscriber) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if subs, ok := pm.channels[channel]; ok {
		delete(subs, sub)
		if len(subs) == 0 {
			delete(pm.channels, channel)
		}
	}
}

func (pm *PubSubManager) Publish(channel string, message []byte) int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	count := 0

	if subs, ok := pm.channels[channel]; ok {
		for sub := range subs {
			if !sub.closed {
				select {
				case sub.ch <- message:
					count++
				default:
				}
			}
		}
	}

	for pattern, subs := range pm.patterns {
		if matchPattern(channel, pattern) {
			for sub := range subs {
				if !sub.closed {
					select {
					case sub.ch <- message:
						count++
					default:
					}
				}
			}
		}
	}

	return count
}

func matchPattern(channel, pattern string) bool {
	if pattern == "" {
		return false
	}
	if pattern == "*" {
		return true
	}
	if len(pattern) > 1 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(channel) >= len(prefix) && channel[:len(prefix)] == prefix
	}
	return channel == pattern
}

func execPublish(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'publish' command")
	}

	channel := string(args[0])
	message := args[1]

	count := GetPubSubManager().Publish(channel, message)
	return &resp.IntegerReply{Num: int64(count)}
}

func execSubscribe(db *DB, args [][]byte, connCh chan []byte) []resp.Reply {
	var replies []resp.Reply

	if len(args) < 1 {
		replies = append(replies, resp.NewErrorReply("ERR wrong number of arguments for 'subscribe' command"))
		return replies
	}

	sub := &Subscriber{
		ch: make(chan []byte, 100),
	}

	for _, arg := range args {
		channel := string(arg)
		GetPubSubManager().Subscribe(channel, sub)
	}

	replies = append(replies, &resp.ArrayReply{
		Replies: []resp.Reply{
			&resp.BulkReply{Arg: []byte("subscribe")},
			&resp.BulkReply{Arg: args[0]},
			&resp.IntegerReply{Num: 1},
		},
	})

	go func() {
		for msg := range sub.ch {
			select {
			case connCh <- msg:
			default:
			}
		}
	}()

	return replies
}

func execPSubscribe(db *DB, args [][]byte, connCh chan []byte) []resp.Reply {
	var replies []resp.Reply

	if len(args) < 1 {
		replies = append(replies, resp.NewErrorReply("ERR wrong number of arguments for 'psubscribe' command"))
		return replies
	}

	sub := &Subscriber{
		ch: make(chan []byte, 100),
	}

	for _, arg := range args {
		pattern := string(arg)
		GetPubSubManager().PSubscribe(pattern, sub)
	}

	replies = append(replies, &resp.ArrayReply{
		Replies: []resp.Reply{
			&resp.BulkReply{Arg: []byte("psubscribe")},
			&resp.BulkReply{Arg: args[0]},
			&resp.IntegerReply{Num: 1},
		},
	})

	go func() {
		for msg := range sub.ch {
			select {
			case connCh <- msg:
			default:
			}
		}
	}()

	return replies
}

func init() {
	RegisterCommand("publish", execPublish, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, nil
		}
		return nil, nil
	}, 3)
}
