// 实现不重复的随机数
package random

import (
	"math/rand"
	"time"

	"golang.org/x/exp/slices"
)

type number interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~int
}

type Random[T number] struct {
	randList []T
}

// range [0, n)
func Number[T number](n int) *Random[T] {
	if n <= 0 {
		return &Random[T]{
			randList: nil,
		}
	}
	list := make([]T, n)
	for i := 0; i < n; i++ {
		list[i] = T(i)
	}
	ra := rand.New(rand.NewSource(time.Now().UnixNano()))
	ra.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
	return &Random[T]{
		randList: list,
	}
}

func (r *Random[T]) Get() T {
	if len(r.randList) == 0 || r.randList == nil {
		return T(-1)
	}
	pickedNumber := r.randList[0]
	r.randList = r.randList[1:]
	return pickedNumber
}

// 删除列表中值为index的元素
func (r *Random[T]) Remove(index T) {
	if index < 0 {
		return
	}
	pos := slices.Index[[]T, T](r.randList, index)
	if pos == -1 {
		return
	}
	r.randList = append(r.randList[:pos], r.randList[pos+1:]...)
}
