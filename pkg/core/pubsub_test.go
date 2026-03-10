package core

import (
	"sync"
	"testing"

	"goflashdb/pkg/resp"
)

func resetPubSubManager() {
	pubSubManager = nil
	pubSubManagerOnce = sync.Once{}
}

func TestPubSubManager_Subscribe(t *testing.T) {
	pm := NewPubSubManager()
	sub := NewSubscriber(1)

	replies := pm.Subscribe(sub, []string{"channel1", "channel2"})

	if len(replies) != 2 {
		t.Errorf("expected 2 replies, got %d", len(replies))
	}

	if len(sub.Channels) != 2 {
		t.Errorf("expected 2 channels, got %d", len(sub.Channels))
	}

	if !sub.Channels["channel1"] || !sub.Channels["channel2"] {
		t.Error("expected channels to be subscribed")
	}
}

func TestPubSubManager_Unsubscribe(t *testing.T) {
	pm := NewPubSubManager()
	sub := NewSubscriber(1)

	pm.Subscribe(sub, []string{"channel1", "channel2"})
	replies := pm.Unsubscribe(sub, []string{"channel1"})

	if len(replies) != 1 {
		t.Errorf("expected 1 reply, got %d", len(replies))
	}

	if len(sub.Channels) != 1 {
		t.Errorf("expected 1 channel remaining, got %d", len(sub.Channels))
	}

	if sub.Channels["channel1"] {
		t.Error("expected channel1 to be unsubscribed")
	}
}

func TestPubSubManager_UnsubscribeAll(t *testing.T) {
	pm := NewPubSubManager()
	sub := NewSubscriber(1)

	pm.Subscribe(sub, []string{"channel1", "channel2"})
	replies := pm.Unsubscribe(sub, []string{})

	if len(replies) != 2 {
		t.Errorf("expected 2 replies, got %d", len(replies))
	}

	if len(sub.Channels) != 0 {
		t.Errorf("expected 0 channels, got %d", len(sub.Channels))
	}
}

func TestPubSubManager_PSubscribe(t *testing.T) {
	pm := NewPubSubManager()
	sub := NewSubscriber(1)

	replies := pm.PSubscribe(sub, []string{"news:*", "sports:*"})

	if len(replies) != 2 {
		t.Errorf("expected 2 replies, got %d", len(replies))
	}

	if len(sub.Patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(sub.Patterns))
	}
}

func TestPubSubManager_PUnsubscribe(t *testing.T) {
	pm := NewPubSubManager()
	sub := NewSubscriber(1)

	pm.PSubscribe(sub, []string{"news:*", "sports:*"})
	replies := pm.PUnsubscribe(sub, []string{"news:*"})

	if len(replies) != 1 {
		t.Errorf("expected 1 reply, got %d", len(replies))
	}

	if len(sub.Patterns) != 1 {
		t.Errorf("expected 1 pattern remaining, got %d", len(sub.Patterns))
	}
}

func TestPubSubManager_Publish(t *testing.T) {
	pm := NewPubSubManager()
	sub1 := NewSubscriber(1)
	sub2 := NewSubscriber(2)

	pm.Subscribe(sub1, []string{"channel1"})
	pm.Subscribe(sub2, []string{"channel1"})

	count := pm.Publish("channel1", []byte("hello"))

	if count != 2 {
		t.Errorf("expected 2 subscribers to receive, got %d", count)
	}

	// Check messages received
	msg1 := <-sub1.MsgCh
	if string(msg1.Payload) != "hello" {
		t.Errorf("expected 'hello', got %s", string(msg1.Payload))
	}

	msg2 := <-sub2.MsgCh
	if string(msg2.Payload) != "hello" {
		t.Errorf("expected 'hello', got %s", string(msg2.Payload))
	}
}

func TestPubSubManager_PublishToPattern(t *testing.T) {
	pm := NewPubSubManager()
	sub := NewSubscriber(1)

	pm.PSubscribe(sub, []string{"news:*"})

	count := pm.Publish("news:sports", []byte("sports news"))

	if count != 1 {
		t.Errorf("expected 1 subscriber to receive, got %d", count)
	}

	msg := <-sub.MsgCh
	if msg.Channel != "news:sports" {
		t.Errorf("expected channel 'news:sports', got %s", msg.Channel)
	}
	if msg.Pattern != "news:*" {
		t.Errorf("expected pattern 'news:*', got %s", msg.Pattern)
	}
}

func TestPubSubManager_Channels(t *testing.T) {
	pm := NewPubSubManager()
	sub := NewSubscriber(1)

	pm.Subscribe(sub, []string{"channel1", "channel2", "test1"})

	channels := pm.Channels("")
	if len(channels) != 3 {
		t.Errorf("expected 3 channels, got %d", len(channels))
	}

	channels = pm.Channels("channel*")
	if len(channels) != 2 {
		t.Errorf("expected 2 channels matching 'channel*', got %d", len(channels))
	}
}

func TestPubSubManager_NumSub(t *testing.T) {
	pm := NewPubSubManager()
	sub1 := NewSubscriber(1)
	sub2 := NewSubscriber(2)

	pm.Subscribe(sub1, []string{"channel1"})
	pm.Subscribe(sub2, []string{"channel1", "channel2"})

	numSub := pm.NumSub([]string{"channel1", "channel2", "channel3"})

	if numSub["channel1"] != 2 {
		t.Errorf("expected 2 subscribers for channel1, got %d", numSub["channel1"])
	}
	if numSub["channel2"] != 1 {
		t.Errorf("expected 1 subscriber for channel2, got %d", numSub["channel2"])
	}
	if numSub["channel3"] != 0 {
		t.Errorf("expected 0 subscribers for channel3, got %d", numSub["channel3"])
	}
}

func TestPubSubManager_NumPat(t *testing.T) {
	pm := NewPubSubManager()
	sub := NewSubscriber(1)

	pm.PSubscribe(sub, []string{"news:*", "sports:*"})

	numPat := pm.NumPat()
	if numPat != 2 {
		t.Errorf("expected 2 pattern subscriptions, got %d", numPat)
	}
}

func TestPubSubManager_RemoveSubscriber(t *testing.T) {
	pm := NewPubSubManager()
	sub := NewSubscriber(1)

	pm.Subscribe(sub, []string{"channel1"})
	pm.PSubscribe(sub, []string{"news:*"})

	pm.RemoveSubscriber(sub)

	if !sub.Closed {
		t.Error("expected subscriber to be closed")
	}

	if len(pm.Channels("")) != 0 {
		t.Errorf("expected 0 channels after removal, got %d", len(pm.Channels("")))
	}

	if pm.NumPat() != 0 {
		t.Errorf("expected 0 patterns after removal, got %d", pm.NumPat())
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		channel  string
		pattern  string
		expected bool
	}{
		{"news:sports", "news:*", true},
		{"news:sports:football", "news:*", true},
		{"news:sports", "news:sports", true},
		{"news:sports", "*", true},
		{"news:sports", "sports:*", false},
		{"news", "new?", true},
		{"news", "n??s", true},
		{"news", "n?s", false},
	}

	for _, tt := range tests {
		result := matchPattern(tt.channel, tt.pattern)
		if result != tt.expected {
			t.Errorf("matchPattern(%q, %q) = %v, expected %v", tt.channel, tt.pattern, result, tt.expected)
		}
	}
}

func TestExecPublish(t *testing.T) {
	resetPubSubManager()
	db := newTestDB()
	pm := GetPubSubManager()
	sub := NewSubscriber(1)
	pm.Subscribe(sub, []string{"mychannel"})

	reply := execPublish(db, [][]byte{[]byte("mychannel"), []byte("hello")})

	intReply, ok := reply.(*resp.IntegerReply)
	if !ok {
		t.Errorf("expected IntegerReply, got %T", reply)
	}
	if intReply.Num != 1 {
		t.Errorf("expected 1, got %d", intReply.Num)
	}
}

func TestExecPublishWrongArgs(t *testing.T) {
	db := newTestDB()

	reply := execPublish(db, [][]byte{[]byte("channel")})

	if !isErrorReply(reply) {
		t.Error("expected error reply for wrong number of arguments")
	}
}

func TestExecPubSub_Channels(t *testing.T) {
	resetPubSubManager()
	db := newTestDB()
	pm := GetPubSubManager()
	sub := NewSubscriber(1)
	pm.Subscribe(sub, []string{"channel1", "channel2"})

	reply := execPubSub(db, [][]byte{[]byte("channels")})

	arrReply, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Errorf("expected ArrayReply, got %T", reply)
	}
	if len(arrReply.Replies) != 2 {
		t.Errorf("expected 2 channels, got %d", len(arrReply.Replies))
	}
}

func TestExecPubSub_NumSub(t *testing.T) {
	resetPubSubManager()
	db := newTestDB()
	pm := GetPubSubManager()
	sub := NewSubscriber(1)
	pm.Subscribe(sub, []string{"channel1"})

	reply := execPubSub(db, [][]byte{[]byte("numsub"), []byte("channel1"), []byte("channel2")})

	arrReply, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Errorf("expected ArrayReply, got %T", reply)
	}
	if len(arrReply.Replies) != 4 {
		t.Errorf("expected 4 elements, got %d", len(arrReply.Replies))
	}
}

func TestExecPubSub_NumPat(t *testing.T) {
	resetPubSubManager()
	db := newTestDB()
	pm := GetPubSubManager()
	sub := NewSubscriber(1)
	pm.PSubscribe(sub, []string{"news:*"})

	reply := execPubSub(db, [][]byte{[]byte("numpat")})

	intReply, ok := reply.(*resp.IntegerReply)
	if !ok {
		t.Errorf("expected IntegerReply, got %T", reply)
	}
	if intReply.Num != 1 {
		t.Errorf("expected 1, got %d", intReply.Num)
	}
}

func TestExecPubSub_UnknownSubcommand(t *testing.T) {
	db := newTestDB()

	reply := execPubSub(db, [][]byte{[]byte("unknown")})

	if !isErrorReply(reply) {
		t.Error("expected error reply for unknown subcommand")
	}
}

func isErrorReply(reply resp.Reply) bool {
	_, ok := reply.(*resp.ErrorReply)
	return ok
}
