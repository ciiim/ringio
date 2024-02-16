package random_test

import (
	"testing"

	"github.com/ciiim/cloudborad/random"
)

func TestRandom(t *testing.T) {
	t.Run("n=-1", func(t *testing.T) {
		n := -1
		list := random.Number[int](n)
		for i := 0; i < n; i++ {
			t.Logf("index:%d,get:%d", i, list.Get())
		}
		end := list.Get()
		if end != -1 {
			t.Log(end)
			t.Error("list should be empty")
		}
		t.Log("end:", end)
	})
	t.Run("n=0", func(t *testing.T) {
		n := 0
		list := random.Number[int](n)
		for i := 0; i < n; i++ {
			t.Logf("index:%d,get:%d", i, list.Get())
		}
		end := list.Get()
		if end != -1 {
			t.Log(end)
			t.Error("list should be empty")
		}
		t.Log("end:", end)
	})
	t.Run("n=10", func(t *testing.T) {
		n := 10
		list := random.Number[int](n)
		for i := 0; i < n; i++ {
			t.Logf("index:%d,get:%d", i, list.Get())
		}
		end := list.Get()
		if end != -1 {
			t.Log(end)
			t.Error("list should be empty")
		}
		t.Log("end:", end)
	})

}
