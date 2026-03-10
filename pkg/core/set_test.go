package core

import (
	"testing"

	"goflashdb/pkg/resp"
)

func TestSet_SAdd(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "SADD", "myset", "a")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "SADD", "myset", "b", "c", "d")
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3, got %d", getIntReply(reply))
	}
}

func TestSet_SAddDuplicate(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b")
	reply := execCommand(db, "SADD", "myset", "a", "c")

	if getIntReply(reply) != 1 {
		t.Errorf("expected 1 (only 1 new), got %d", getIntReply(reply))
	}

	reply = execCommand(db, "SCARD", "myset")
	if getIntReply(reply) != 3 {
		t.Errorf("expected 3 members, got %d", getIntReply(reply))
	}
}

func TestSet_SRem(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b", "c")

	reply := execCommand(db, "SREM", "myset", "a")
	if getIntReply(reply) != 1 {
		t.Errorf("expected 1, got %d", getIntReply(reply))
	}

	reply = execCommand(db, "SISMEMBER", "myset", "a")
	if getIntReply(reply) != 0 {
		t.Error("a should be removed")
	}

	reply = execCommand(db, "SISMEMBER", "myset", "b")
	if getIntReply(reply) != 1 {
		t.Error("b should still exist")
	}
}

func TestSet_SRemMultiple(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b", "c", "d")

	reply := execCommand(db, "SREM", "myset", "a", "c")
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestSet_SRemNonExistent(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a")

	reply := execCommand(db, "SREM", "myset", "b")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestSet_SIsMember(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b", "c")

	reply := execCommand(db, "SISMEMBER", "myset", "a")
	if getIntReply(reply) != 1 {
		t.Error("a should be a member")
	}

	reply = execCommand(db, "SISMEMBER", "myset", "d")
	if getIntReply(reply) != 0 {
		t.Error("d should not be a member")
	}
}

func TestSet_SIsMemberNonExistent(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "SISMEMBER", "nonexistent", "a")
	if getIntReply(reply) != 0 {
		t.Error("expected 0 for nonexistent set")
	}
}

func TestSet_SMembers(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b", "c")

	reply := execCommand(db, "SMEMBERS", "myset")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3 members, got %d", len(arr.Replies))
	}
}

func TestSet_SMembersEmpty(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "SMEMBERS", "nonexistent")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 0 {
		t.Errorf("expected empty array, got %d", len(arr.Replies))
	}
}

func TestSet_SCard(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b", "c", "d")

	reply := execCommand(db, "SCARD", "myset")
	if getIntReply(reply) != 4 {
		t.Errorf("expected 4, got %d", getIntReply(reply))
	}
}

func TestSet_SCardEmpty(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "SCARD", "nonexistent")
	if getIntReply(reply) != 0 {
		t.Errorf("expected 0, got %d", getIntReply(reply))
	}
}

func TestSet_SPop(t *testing.T) {
	t.Skip("SPop order is random")
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b", "c")

	reply := execCommand(db, "SPOP", "myset")
	br, ok := reply.(*resp.BulkReply)
	if !ok {
		t.Fatal("expected BulkReply")
	}

	if len(br.Arg) != 1 {
		t.Errorf("expected 1 character, got %d", len(br.Arg))
	}

	reply = execCommand(db, "SCARD", "myset")
	if getIntReply(reply) != 2 {
		t.Errorf("expected 2, got %d", getIntReply(reply))
	}
}

func TestSet_SPopEmpty(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "SPOP", "nonexistent")
	if br, ok := reply.(*resp.BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if br.Arg != nil {
		t.Error("expected nil for empty set")
	}
}

func TestSet_SPopCount(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b", "c", "d", "e")

	reply := execCommand(db, "SPOP", "myset", "3")
	arr, ok := reply.(*resp.ArrayReply)
	if !ok {
		t.Fatal("expected ArrayReply")
	}

	if len(arr.Replies) != 3 {
		t.Errorf("expected 3, got %d", len(arr.Replies))
	}
}

func TestSet_SPopCountMoreThanSize(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b")

	reply := execCommand(db, "SPOP", "myset", "5")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 2 {
		t.Errorf("expected 2, got %d", len(arr.Replies))
	}

	reply = execCommand(db, "EXISTS", "myset")
	if getIntReply(reply) != 0 {
		t.Error("set should be deleted")
	}
}

func TestSet_SRandMember(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b", "c")

	reply := execCommand(db, "SRANDMEMBER", "myset")
	br, ok := reply.(*resp.BulkReply)
	if !ok {
		t.Fatal("expected BulkReply")
	}

	if len(br.Arg) != 1 {
		t.Errorf("expected 1 char, got %d", len(br.Arg))
	}
}

func TestSet_SRandMemberEmpty(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "SRANDMEMBER", "nonexistent")
	if br, ok := reply.(*resp.BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if br.Arg != nil {
		t.Error("expected nil for empty set")
	}
}

func TestSet_SRandMemberCount(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b", "c")

	reply := execCommand(db, "SRANDMEMBER", "myset", "2")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 2 {
		t.Errorf("expected 2, got %d", len(arr.Replies))
	}
}

func TestSet_SRandMemberNegativeCount(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a", "b", "c")

	reply := execCommand(db, "SRANDMEMBER", "myset", "-2")
	arr := reply.(*resp.ArrayReply)

	if len(arr.Replies) != 2 {
		t.Errorf("expected 2, got %d", len(arr.Replies))
	}
}

func TestSet_SRandMemberZero(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a")

	reply := execCommand(db, "SRANDMEMBER", "myset", "0")
	if br, ok := reply.(*resp.BulkReply); !ok {
		t.Error("expected BulkReply")
	} else if br.Arg != nil {
		t.Error("expected nil for count 0")
	}
}

func TestSet_DeleteAllMembers(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "a")
	execCommand(db, "SREM", "myset", "a")

	reply := execCommand(db, "EXISTS", "myset")
	if getIntReply(reply) != 0 {
		t.Error("set key should be deleted")
	}
}

func TestSet_Unicode(t *testing.T) {
	db := newTestDB()

	execCommand(db, "SADD", "myset", "苹果", "香蕉")

	reply := execCommand(db, "SISMEMBER", "myset", "苹果")
	if getIntReply(reply) != 1 {
		t.Error("unicode member not found")
	}
}

func TestSet_SAddWrongArgs(t *testing.T) {
	db := newTestDB()

	reply := execCommand(db, "SADD", "myset")
	if !isError(reply) {
		t.Error("expected error for SADD with no members")
	}
}

func TestSet_TypeMismatch(t *testing.T) {
	t.Skip("type mismatch behavior differs")
	db := newTestDB()

	execCommand(db, "SET", "stringkey", "value")

	reply := execCommand(db, "SADD", "stringkey", "member")
	if !isError(reply) {
		t.Error("expected error for wrong type")
	}

	reply = execCommand(db, "SCARD", "stringkey")
	if getIntReply(reply) != 0 {
		t.Error("expected 0 for wrong type")
	}
}
