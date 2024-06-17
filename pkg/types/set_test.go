package types

import "testing"

func TestSet(t *testing.T) {
	s1 := NewSet()
	s1.Add(1)
	s1.Add(2)
	s1.Add(3)
	t.Log("Set s1", s1.Elements())
	s1.Add(2)
	t.Log("Set s1", s1.Elements())
	t.Log("Contains 3", s1.Contains(3))
	t.Log("Contains 5", s1.Contains(5))

	s2 := NewSet()
	s2.Add("hello")
	s2.Add("world")
	t.Log("Set s2", s2.Elements())
	t.Log("Contains hello", s2.Contains("hello"))
	s2.Remove("world")
	t.Log("Set s2", s2.Elements())

	s3 := NewSet(1, 2, 3)
	s3.Add("hello", "world")
	t.Log("Set s3", s3.Elements())

	s4 := NewSet(1, 2, 3)
	s4.Add(1, 3, 5)
	s4.Add(2, 4, 7)
	t.Log("Set s4", s4.Elements())
	array := make([]int, 0, s4.Size())
	for _, item := range s4.Elements() {
		array = append(array, item.(int))
	}
	t.Log("Convert s4 to array", array)

	s5 := NewSet(1, 2, 3)
	var a2 = [3]int{4, 5, 6}
	s5.Add(a2)
	t.Log("slice array[:]", a2[:])
	t.Log("Set s5", s5.Elements(), s5.Size())
}
