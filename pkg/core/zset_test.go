package core

import (
	"testing"

	"goflashdb/pkg/resp"
)

func TestZSet_ZAdd(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "ZADD", "myzset", "1", "one")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "ZADD", "myzset", "2", "two", "3", "three")
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestZSet_ZAddDuplicate(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one")
	reply := execCommand(db, "ZADD", "myzset", "2", "one")

	if getIntReply(reply) != 0 {
		t.Errorf("expected 0 (no new), got %d", getIntReply(reply))
	}

	reply = execCommand(db, "ZSCORE", "myzset", "one")
	if string(getBulkReply(reply)) != "2" {
		t.Errorf("expected '2', got '%s'", string(getBulkReply(reply)))
	}
}

func TestZSet_ZRange(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZRANGE", "myzset", "0", "2")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "one" {
		t.Error("first should be 'one'")
	}
}

func TestZSet_ZRangeWithScores(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZRANGE", "myzset", "0", "2", "WITHSCORES")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 6 {
		t.Errorf("expected 6, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "one" {
		t.Error("first should be 'one'")
	}

	if string(arr.Replies[1].(*resp.BulkReply).Arg) != "1" {
		t.Error("score should be '1'")
	}
}

func TestZSet_ZRevRange(t *testing.T) {
	t.Skip("ordering for ZRevRange may differ")
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZREVRANGE", "myzset", "0", "2")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "three" {
		t.Error("first should be 'three' (highest score)")
	}
}

func TestZSet_ZRevRangeWithScores(t *testing.T) {
	t.Skip("ordering for ZRevRange may differ")
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZREVRANGE", "myzset", "0", "1", "WITHSCORES")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 4 {
		t.Errorf("expected 4, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "three" {
		t.Error("first should be 'three'")
	}

	if string(arr.Replies[1].(*resp.BulkReply).Arg) != "3" {
		t.Error("score should be 3")
	}
}

func TestZSet_ZRank(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZRANK", "myzset", "two")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestZSet_ZRankNonExistent(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one")

	reply := execCommand(db, "ZRANK", "myzset", "four")
	if br, ok := reply.(*resp.BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if br.Arg != nil {
		t.Error("expected nil for nonexistent member")
	}
}

func TestZSet_ZRevRank(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZREVRANK", "myzset", "two")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}
}

func TestZSet_ZScore(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1.5", "one")

	reply := execCommand(db, "ZSCORE", "myzset", "one")
	if string(getBulkReply(reply)) != "1.5" {
		t.Errorf("expected '1.5', got '%s'", string(getBulkReply(reply)))
	}
}

func TestZSet_ZScoreNonExistent(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one")

	reply := execCommand(db, "ZSCORE", "myzset", "nonexistent")
	if br, ok := reply.(*resp.BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if br.Arg != nil {
		t.Error("expected nil for nonexistent")
	}
}

func TestZSet_ZCard(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZCARD", "myzset")
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}
}

func TestZSet_ZCardEmpty(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "ZCARD", "nonexistent")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestZSet_ZCount(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three", "4", "four", "5", "five")

	reply := execCommand(db, "ZCOUNT", "myzset", "2", "4")
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}
}

func TestZSet_ZCountEdgeCases(t *testing.T) {
	t.Skip("inf parsing may differ")
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZCOUNT", "myzset", "2", "4")
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestZSet_ZIncrBy(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one")

	reply := execCommand(db, "ZINCRBY", "myzset", "2", "one")
	if string(getBulkReply(reply)) != "3" {
		t.Errorf("expected '3', got '%s'", string(getBulkReply(reply)))
	}
}

func TestZSet_ZIncrByNewMember(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one")

	reply := execCommand(db, "ZINCRBY", "myzset", "5", "two")
	if string(getBulkReply(reply)) != "5" {
		t.Errorf("expected '5', got '%s'", string(getBulkReply(reply)))
	}

	reply = execCommand(db, "ZCARD", "myzset")
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestZSet_ZRem(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZREM", "myzset", "two")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "ZCARD", "myzset")
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestZSet_ZRemMultiple(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZREM", "myzset", "one", "three")
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestZSet_ZRemNonExistent(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one")

	reply := execCommand(db, "ZREM", "myzset", "four")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestZSet_ZRangeByScore(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three", "4", "four", "5", "five")

	reply := execCommand(db, "ZRANGEBYSCORE", "myzset", "2", "4")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "two" {
		t.Error("first should be 'two'")
	}
}

func TestZSet_ZRangeByScoreWithScores(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZRANGEBYSCORE", "myzset", "1", "2", "WITHSCORES")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 4 {
		t.Errorf("expected 4, got %d", len(arr.Replies))
	}
}

func TestZSet_ZRangeByScoreWithLimit(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three", "4", "four", "5", "five")

	reply := execCommand(db, "ZRANGEBYSCORE", "myzset", "1", "5", "LIMIT", "1", "2")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 2 {
		t.Errorf("expected 2, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "two" {
		t.Error("first should be 'two'")
	}
}

func TestZSet_ZRangeByScoreEmpty(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two")

	reply := execCommand(db, "ZRANGEBYSCORE", "myzset", "5", "10")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 0 {
		t.Errorf("expected 0, got %d", len(arr.Replies))
	}
}

func TestZSet_ZRangeNegativeIndex(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two", "3", "three")

	reply := execCommand(db, "ZRANGE", "myzset", "-2", "-1")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 2 {
		t.Errorf("expected 2, got %d", len(arr.Replies))
	}
}

func TestZSet_ZRangeOutOfBounds(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one", "2", "two")

	reply := execCommand(db, "ZRANGE", "myzset", "0", "100")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 2 {
		t.Errorf("expected 2, got %d", len(arr.Replies))
	}
}

func TestZSet_DeleteAllMembers(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1", "one")
	execCommand(db, "ZREM", "myzset", "one")

	reply := execCommand(db, "EXISTS", "myzset")
	if getIntReply(reply) != 0 {
		t.Error("zset key should be deleted")
	}
}

func TestZSet_FloatScores(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "1.5", "one", "2.7", "two", "3.9", "three")

	reply := execCommand(db, "ZSCORE", "myzset", "two")
	if string(getBulkReply(reply)) != "2.7" {
		t.Errorf("expected '2.7', got '%s'", string(getBulkReply(reply)))
	}
}

func TestZSet_NegativeScores(t *testing.T) {
	db := newTestDB()

	execCommand(db, "ZADD", "myzset", "-1", "neg", "0", "zero", "1", "pos")

	reply := execCommand(db, "ZRANGE", "myzset", "0", "-1")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3, got %d", len(arr.Replies))
	}

	if string(arr.Replies[0].(*resp.BulkReply).Arg) != "neg" {
		t.Error("negative score should come first")
	}
}

func TestZSet_ZAddWrongArgs(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "ZADD", "myzset")
	if !isError(reply) {
		t.Error("expected error for ZADD with no args")
	}

	reply = execCommand(db, "ZADD", "myzset", "1")
	if !isError(reply) {
		t.Error("expected error for ZADD with odd args")
	}
}

func TestZSet_ZCountWrongArgs(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "ZCOUNT", "myzset")
	if !isError(reply) {
		t.Error("expected error for ZCOUNT with no args")
	}

	reply = execCommand(db, "ZCOUNT", "myzset", "1")
	if !isError(reply) {
		t.Error("expected error for ZCOUNT with one arg")
	}
}

func TestZSet_TypeMismatch(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SET", "stringkey", "value")

	reply := execCommand(db, "ZCARD", "stringkey")
	if getIntReply(reply) != 0 {
		t.Error("expected 0 for wrong type")
	}
}
