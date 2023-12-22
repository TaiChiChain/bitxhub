package snapshot

import (
	"fmt"
)

const (
	snapshotKey = "snap-"
)

func CompositeSnapAccountKey(addr string) []byte {
	return append([]byte(snapshotKey), []byte(addr)...)
}

func CompositeSnapStorageKey(addr string, key []byte) []byte {
	res := append([]byte(snapshotKey), []byte(addr)...)
	res = append(res, key...)
	return res
}

func CompositeSnapJournalKey(value any) []byte {
	return append([]byte(snapshotKey), []byte(fmt.Sprintf("%v", value))...)
}
