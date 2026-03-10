package core

import (
	"testing"
)

func TestNewHashData(t *testing.T) {
	hashData := NewHashData()
	if hashData == nil {
		t.Error("NewHashData should not return nil")
	}
	if hashData.data == nil {
		t.Error("data should not be nil")
	}
	if hashData.expireAt != 0 {
		t.Error("expireAt should be 0")
	}
}

func TestNewListData(t *testing.T) {
	listData := NewListData()
	if listData == nil {
		t.Error("NewListData should not return nil")
	}
	if listData.items == nil {
		t.Error("items should not be nil")
	}
	if listData.expireAt != 0 {
		t.Error("expireAt should be 0")
	}
}

func TestNewSetData(t *testing.T) {
	setData := NewSetData()
	if setData == nil {
		t.Error("NewSetData should not return nil")
	}
	if setData.members == nil {
		t.Error("members should not be nil")
	}
	if setData.expireAt != 0 {
		t.Error("expireAt should be 0")
	}
}

func TestNewZSetData(t *testing.T) {
	zsetData := NewZSetData()
	if zsetData == nil {
		t.Error("NewZSetData should not return nil")
	}
	if zsetData.members == nil {
		t.Error("members should not be nil")
	}
	if zsetData.scoreMap == nil {
		t.Error("scoreMap should not be nil")
	}
	if zsetData.expireAt != 0 {
		t.Error("expireAt should be 0")
	}
}

func TestZSetDataAdd(t *testing.T) {
	zsetData := NewZSetData()

	existed := zsetData.Add(1.0, []byte("member1"))
	if !existed {
		t.Error("Expected true for new member")
	}

	existed = zsetData.Add(2.0, []byte("member1"))
	if existed {
		t.Error("Expected false for existing member")
	}

	if len(zsetData.members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(zsetData.members))
	}
}

func TestZSetDataAddMultiple(t *testing.T) {
	zsetData := NewZSetData()

	zsetData.Add(3.0, []byte("member3"))
	zsetData.Add(1.0, []byte("member1"))
	zsetData.Add(2.0, []byte("member2"))

	if len(zsetData.members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(zsetData.members))
	}

	if zsetData.members[0].score != 1.0 {
		t.Errorf("Expected first member to have score 1.0, got %f", zsetData.members[0].score)
	}
}
