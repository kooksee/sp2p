// Copyright 2017 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// RandId return a rand string used in frp.
func RandId() (id string, err error) {
	return RandIdWithLen(8)
}

// RandIdWithLen return a rand string with idLen length.
func RandIdWithLen(idLen int) (id string, err error) {
	b := make([]byte, idLen)
	_, err = rand.Read(b)
	if err != nil {
		return
	}

	id = fmt.Sprintf("%x", b)
	return
}

func GetAuthKey(token string, timestamp int64) (key string) {
	token = token + fmt.Sprintf("%d", timestamp)
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(token))
	data := md5Ctx.Sum(nil)
	return hex.EncodeToString(data)
}

// for example: rangeStr is "1000-2000,2001,2002,3000-4000", return an array as port ranges.
func GetPortRanges(rangeStr string) (portRanges [][2]int64, err error) {
	// for example: 1000-2000,2001,2002,3000-4000
	rangeArray := strings.Split(rangeStr, ",")
	for _, portRangeStr := range rangeArray {
		// 1000-2000 or 2001
		portArray := strings.Split(portRangeStr, "-")
		// length: only 1 or 2 is correct
		rangeType := len(portArray)
		if rangeType == 1 {
			singlePort, err := strconv.ParseInt(portArray[0], 10, 64)
			if err != nil {
				return [][2]int64{}, err
			}
			portRanges = append(portRanges, [2]int64{singlePort, singlePort})
		} else if rangeType == 2 {
			min, err := strconv.ParseInt(portArray[0], 10, 64)
			if err != nil {
				return [][2]int64{}, err
			}
			max, err := strconv.ParseInt(portArray[1], 10, 64)
			if err != nil {
				return [][2]int64{}, err
			}
			if max < min {
				return [][2]int64{}, fmt.Errorf("range incorrect")
			}
			portRanges = append(portRanges, [2]int64{min, max})
		} else {
			return [][2]int64{}, fmt.Errorf("format error")
		}
	}
	return portRanges, nil
}

func ContainsPort(portRanges [][2]int64, port int64) bool {
	for _, pr := range portRanges {
		if port >= pr[0] && port <= pr[1] {
			return true
		}
	}
	return false
}

func PortRangesCut(portRanges [][2]int64, port int64) [][2]int64 {
	var tmpRanges [][2]int64
	for _, pr := range portRanges {
		if port >= pr[0] && port <= pr[1] {
			leftRange := [2]int64{pr[0], port - 1}
			rightRange := [2]int64{port + 1, pr[1]}
			if leftRange[0] <= leftRange[1] {
				tmpRanges = append(tmpRanges, leftRange)
			}
			if rightRange[0] <= rightRange[1] {
				tmpRanges = append(tmpRanges, rightRange)
			}
		} else {
			tmpRanges = append(tmpRanges, pr)
		}
	}
	return tmpRanges
}

func IsFileExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// 范围判断 min <= v <= max
func between(v, min, max []byte) bool {
	return bytes.Compare(v, min) >= 0 && bytes.Compare(v, max) <= 0
}

// 复制数组
func copyBytes(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

// 使用二进制存储整形
func IntToByte(x int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(x))
	return b
}

func ByteToInt(x []byte) int {
	return int(binary.BigEndian.Uint64(x))
}

func f(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

// S2b converts string to a byte slice without memory allocation.
// "abc" -> []byte("abc")
func S2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// B2s converts byte slice to a string without memory allocation.
// []byte("abc") -> "abc" s
func B2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// B2ds return a Digit string of v
// v (8-byte big endian) -> uint64(123456) -> "123456".
func B2ds(v []byte) string {
	return strconv.FormatUint(binary.BigEndian.Uint64(v), 10)
}

// Btoi return an int64 of v
// v (8-byte big endian) -> uint64(123456).
func B2i(v []byte) uint64 {
	return binary.BigEndian.Uint64(v)
}

// DS2i returns uint64 of Digit string
// v ("123456") -> uint64(123456).
func DS2i(v string) uint64 {
	i, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return uint64(0)
	}
	return i
}

// Itob returns an 8-byte big endian representation of v
// v uint64(123456) -> 8-byte big endian.
func I2b(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// DS2b returns an 8-byte big endian representation of Digit string
// v ("123456") -> uint64(123456) -> 8-byte big endian.
func DS2b(v string) []byte {
	i, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return []byte("")
	}
	return I2b(i)
}

// BConcat concat a list of byte
func BConcat(slices ... []byte) []byte {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	tmp := make([]byte, totalLen)
	var i int
	for _, s := range slices {
		i += copy(tmp[i:], s)
	}
	return tmp
}

func BMap(m [][]byte, fn func(i int, k []byte) []byte) [][]byte {
	for i, d := range m {
		m[i] = fn(i, d)
	}
	return m
}

// 生成count个[start,end)结束的不重复的随机数
func GenRandom(start int, end int, count int) map[int]bool {

	// 范围检查
	if end < start || (end-start) < count {
		return nil
	}

	nums := map[int]bool{}

	// 随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; len(nums) < count && i < count; {

		// 生成随机数
		num := r.Intn(end-start) + start
		if nums[num] {
			continue
		}

		i++
		nums[num] = true
	}

	return nums
}
