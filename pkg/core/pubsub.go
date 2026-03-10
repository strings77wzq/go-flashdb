package core

import (
	"path/filepath"
	"sync"

	"goflashdb/pkg/resp"
)

var (
	pubSubManager     *PubSubManager
	pubSubManagerOnce sync.Once
)

// PubSubManager manages pub/sub subscriptions
type PubSubManager struct {
	mu       sync.RWMutex
	channels map[string]map[*Subscriber]bool // channel -> subscribers
	patterns map[string]map[*Subscriber]bool // pattern -> subscribers
}

// Subscriber represents a client subscription
type Subscriber struct {
	ID       int64
	Channels map[string]bool // subscribed channels
	Patterns map[string]bool // subscribed patterns
	MsgCh    chan *PubSubMessage
	Closed   bool
	mu       sync.RWMutex
}

// PubSubMessage represents a pub/sub message
type PubSubMessage struct {
	Channel   string
	Pattern   string // empty for regular subscribe
	Payload   []byte
	IsMessage bool // true for message, false for subscribe/unsubscribe notification
}

// NewPubSubManager creates a new PubSubManager
func NewPubSubManager() *PubSubManager {
	return &PubSubManager{
		channels: make(map[string]map[*Subscriber]bool),
		patterns: make(map[string]map[*Subscriber]bool),
	}
}

// GetPubSubManager returns the global PubSubManager singleton
func GetPubSubManager() *PubSubManager {
	pubSubManagerOnce.Do(func() {
		pubSubManager = NewPubSubManager()
	})
	return pubSubManager
}

// NewSubscriber creates a new subscriber
func NewSubscriber(id int64) *Subscriber {
	return &Subscriber{
		ID:       id,
		Channels: make(map[string]bool),
		Patterns: make(map[string]bool),
		MsgCh:    make(chan *PubSubMessage, 256),
		Closed:   false,
	}
}

// Subscribe subscribes to one or more channels
func (pm *PubSubManager) Subscribe(sub *Subscriber, channels []string) []resp.Reply {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var replies []resp.Reply
	count := len(sub.Channels)

	for _, channel := range channels {
		if pm.channels[channel] == nil {
			pm.channels[channel] = make(map[*Subscriber]bool)
		}
		pm.channels[channel][sub] = true
		sub.Channels[channel] = true
		count++

		replies = append(replies, &resp.ArrayReply{
			Replies: []resp.Reply{
				&resp.BulkReply{Arg: []byte("subscribe")},
				&resp.BulkReply{Arg: []byte(channel)},
				&resp.IntegerReply{Num: int64(count)},
			},
		})
	}

	return replies
}

// Unsubscribe unsubscribes from one or more channels
func (pm *PubSubManager) Unsubscribe(sub *Subscriber, channels []string) []resp.Reply {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var replies []resp.Reply

	// If no channels specified, unsubscribe from all
	if len(channels) == 0 {
		for channel := range sub.Channels {
			channels = append(channels, channel)
		}
	}

	for _, channel := range channels {
		if subs, ok := pm.channels[channel]; ok {
			delete(subs, sub)
			if len(subs) == 0 {
				delete(pm.channels, channel)
			}
		}
		delete(sub.Channels, channel)

		replies = append(replies, &resp.ArrayReply{
			Replies: []resp.Reply{
				&resp.BulkReply{Arg: []byte("unsubscribe")},
				&resp.BulkReply{Arg: []byte(channel)},
				&resp.IntegerReply{Num: int64(len(sub.Channels) + len(sub.Patterns))},
			},
		})
	}

	return replies
}

// PSubscribe subscribes to one or more patterns
func (pm *PubSubManager) PSubscribe(sub *Subscriber, patterns []string) []resp.Reply {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var replies []resp.Reply
	count := len(sub.Channels) + len(sub.Patterns)

	for _, pattern := range patterns {
		if pm.patterns[pattern] == nil {
			pm.patterns[pattern] = make(map[*Subscriber]bool)
		}
		pm.patterns[pattern][sub] = true
		sub.Patterns[pattern] = true
		count++

		replies = append(replies, &resp.ArrayReply{
			Replies: []resp.Reply{
				&resp.BulkReply{Arg: []byte("psubscribe")},
				&resp.BulkReply{Arg: []byte(pattern)},
				&resp.IntegerReply{Num: int64(count)},
			},
		})
	}

	return replies
}

// PUnsubscribe unsubscribes from one or more patterns
func (pm *PubSubManager) PUnsubscribe(sub *Subscriber, patterns []string) []resp.Reply {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var replies []resp.Reply

	// If no patterns specified, unsubscribe from all
	if len(patterns) == 0 {
		for pattern := range sub.Patterns {
			patterns = append(patterns, pattern)
		}
	}

	for _, pattern := range patterns {
		if subs, ok := pm.patterns[pattern]; ok {
			delete(subs, sub)
			if len(subs) == 0 {
				delete(pm.patterns, pattern)
			}
		}
		delete(sub.Patterns, pattern)

		replies = append(replies, &resp.ArrayReply{
			Replies: []resp.Reply{
				&resp.BulkReply{Arg: []byte("punsubscribe")},
				&resp.BulkReply{Arg: []byte(pattern)},
				&resp.IntegerReply{Num: int64(len(sub.Channels) + len(sub.Patterns))},
			},
		})
	}

	return replies
}

// Publish publishes a message to a channel
func (pm *PubSubManager) Publish(channel string, message []byte) int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	count := 0

	// Send to direct channel subscribers
	if subs, ok := pm.channels[channel]; ok {
		for sub := range subs {
			sub.mu.RLock()
			closed := sub.Closed
			sub.mu.RUnlock()

			if !closed {
				select {
				case sub.MsgCh <- &PubSubMessage{
					Channel:   channel,
					Payload:   message,
					IsMessage: true,
				}:
					count++
				default:
					// Channel full, skip
				}
			}
		}
	}

	// Send to pattern subscribers
	for pattern, subs := range pm.patterns {
		if matchPattern(channel, pattern) {
			for sub := range subs {
				sub.mu.RLock()
				closed := sub.Closed
				sub.mu.RUnlock()

				if !closed {
					select {
					case sub.MsgCh <- &PubSubMessage{
						Channel:   channel,
						Pattern:   pattern,
						Payload:   message,
						IsMessage: true,
					}:
						count++
					default:
						// Channel full, skip
					}
				}
			}
		}
	}

	return count
}

// Channels returns list of active channels matching pattern
func (pm *PubSubManager) Channels(pattern string) []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var result []string
	for channel := range pm.channels {
		if len(pm.channels[channel]) > 0 {
			if pattern == "" || matchPattern(channel, pattern) {
				result = append(result, channel)
			}
		}
	}
	return result
}

// NumSub returns the number of subscribers for each channel
func (pm *PubSubManager) NumSub(channels []string) map[string]int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]int)
	for _, channel := range channels {
		if subs, ok := pm.channels[channel]; ok {
			result[channel] = len(subs)
		} else {
			result[channel] = 0
		}
	}
	return result
}

// NumPat returns the number of pattern subscriptions
func (pm *PubSubManager) NumPat() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	count := 0
	for _, subs := range pm.patterns {
		count += len(subs)
	}
	return count
}

// RemoveSubscriber removes a subscriber from all subscriptions
func (pm *PubSubManager) RemoveSubscriber(sub *Subscriber) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	sub.mu.Lock()
	sub.Closed = true
	close(sub.MsgCh)
	sub.mu.Unlock()

	// Remove from channels
	for channel := range sub.Channels {
		if subs, ok := pm.channels[channel]; ok {
			delete(subs, sub)
			if len(subs) == 0 {
				delete(pm.channels, channel)
			}
		}
	}

	// Remove from patterns
	for pattern := range sub.Patterns {
		if subs, ok := pm.patterns[pattern]; ok {
			delete(subs, sub)
			if len(subs) == 0 {
				delete(pm.patterns, pattern)
			}
		}
	}
}

// matchPattern checks if channel matches the pattern
// Supports glob-style patterns: * matches any sequence, ? matches single char
func matchPattern(channel, pattern string) bool {
	if pattern == "" {
		return false
	}
	if pattern == "*" {
		return true
	}

	matched, _ := filepath.Match(pattern, channel)
	return matched
}

// execPublish executes PUBLISH command
func execPublish(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'publish' command")
	}

	channel := string(args[0])
	message := args[1]

	count := GetPubSubManager().Publish(channel, message)
	return &resp.IntegerReply{Num: int64(count)}
}

// execPubSub executes PUBSUB command
func execPubSub(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'pubsub' command")
	}

	subcommand := string(args[0])
	pm := GetPubSubManager()

	switch subcommand {
	case "channels":
		pattern := ""
		if len(args) > 1 {
			pattern = string(args[1])
		}
		channels := pm.Channels(pattern)
		replies := make([]resp.Reply, len(channels))
		for i, ch := range channels {
			replies[i] = &resp.BulkReply{Arg: []byte(ch)}
		}
		return &resp.ArrayReply{Replies: replies}

	case "numsub":
		if len(args) < 2 {
			return resp.NewErrorReply("ERR wrong number of arguments for 'pubsub|numsub' command")
		}
		channels := make([]string, len(args)-1)
		for i := 1; i < len(args); i++ {
			channels[i-1] = string(args[i])
		}
		numSub := pm.NumSub(channels)

		var replies []resp.Reply
		for _, ch := range channels {
			replies = append(replies,
				&resp.BulkReply{Arg: []byte(ch)},
				&resp.IntegerReply{Num: int64(numSub[ch])},
			)
		}
		return &resp.ArrayReply{Replies: replies}

	case "numpat":
		return &resp.IntegerReply{Num: int64(pm.NumPat())}

	default:
		return resp.NewErrorReply("ERR unknown subcommand '" + subcommand + "'")
	}
}

func init() {
	RegisterCommand("publish", execPublish, nil, 3)
	RegisterCommand("pubsub", execPubSub, nil, -2)
}
