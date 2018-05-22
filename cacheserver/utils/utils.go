package utils

import (
	"hash/crc32"
	"unsafe"
)

func BucketHash(b []byte) int {
	return int(crc32.ChecksumIEEE(b))
}

func GetBucketIndex(bucketName []byte, bucketNum int) int {
	return BucketHash(bucketName) % bucketNum
}

func BytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
