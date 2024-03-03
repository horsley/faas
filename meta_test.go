package main

import "testing"

func Test_pathMeta(t *testing.T) {
	k1, _ := MetaOf("/").WriteKey()
	k2, _ := MetaOf("/some/not/exist").WriteKey()
	if k1 != k2 {
		t.Error("not inherit key from parent")
	}

	_, ok := MetaOf("../../../some/not/exist").WriteKey()
	if ok {
		t.Error("path travel out of root")
	}

	if MetaOf("/some/sub/item").SetWriteKey("123") != nil {
		t.Error("set key err")
	}

	if k3, _ := MetaOf("/some/sub/item").WriteKey(); k3 != "123" {
		t.Error("key not match")
	}

	if MetaOf("/some/sub/item/key").SetWriteKey("678") == nil {
		t.Error("write sub item key of existed item")
	}

	if err := MetaOf("/some/sub/item").Destroy(); err != nil {
		t.Error("destroy err", err)
	}
}
