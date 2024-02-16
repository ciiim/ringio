package hashchunk_test

import (
	"strings"
)

var testFileName = "test.txt"
var testData = strings.Repeat("A", 1024*5)
var testDataLen = len(testData)
var testCap int64 = 1024 * 1024 * 1024
var testChunkSize int64 = 1024 * 1024 * 4 //4MB
