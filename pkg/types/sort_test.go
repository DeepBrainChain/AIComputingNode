package types

import (
	"slices"
	"sort"
	"testing"
)

// https://pkg.go.dev/sort
// https://stackoverflow.com/questions/37695209/golang-sort-slice-ascending-or-descending

func TestUnorderedMap(t *testing.T) {
	map1 := make(map[int]uint32)
	map1[24] = 240
	map1[17] = 170
	map1[9] = 90
	map1[11] = 110
	map1[55] = 550
	t.Log("First iterate")
	for key, value := range map1 {
		t.Logf("{%v, %v}", key, value)
	}
	t.Log("Second iterate")
	for key, value := range map1 {
		t.Logf("{%v, %v}", key, value)
	}
	t.Log("The order of each traversal may be different")
}

func TestOrderedMapByKey(t *testing.T) {
	map1 := map[int]uint32{}
	keys := []int{}
	map1[24] = 240
	map1[17] = 170
	map1[9] = 90
	map1[11] = 110
	map1[55] = 550
	t.Log("Origin map")
	for key, value := range map1 {
		t.Logf("{%v, %v}", key, value)
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	t.Log("After sort.Slice func(i, j int) bool { return i < j }")
	for _, key := range keys {
		t.Logf("{%v, %v}", key, map1[key])
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})
	t.Log("After sort.Slice func(i, j int) bool { return i > j }")
	for _, key := range keys {
		t.Logf("{%v, %v}", key, map1[key])
	}

	sort.Ints(keys)
	t.Log("After sort.Ints")
	for _, key := range keys {
		t.Logf("{%v, %v}", key, map1[key])
	}

	// sort.Sort(sort.IntSlice(keys))
	slices.Reverse(keys)
	t.Log("After slices.Reverse")
	for _, key := range keys {
		t.Logf("{%v, %v}", key, map1[key])
	}

	slices.Sort(keys)
	t.Log("After slices.Sort")
	for _, key := range keys {
		t.Logf("{%v, %v}", key, map1[key])
	}
}

func TestOrderedMapByValue(t *testing.T) {
	map1 := map[string]uint32{}
	persons := []testPerson{}

	map1["C/C++"] = 240
	map1["Java"] = 170
	map1["Rust"] = 90
	map1["Python"] = 110
	map1["Golang"] = 550

	t.Log("Origin map")
	for key, value := range map1 {
		t.Logf("{%v, %v}", key, value)
		persons = append(persons, testPerson{
			Name: key,
			Age:  int(value),
		})
	}

	t.Log("Only maps with integer key values can be sorted.")
	t.Log("Convert to integer key values and then sort")
	sort.Slice(persons, func(i, j int) bool {
		return persons[i].Age < persons[j].Age
	})
	for _, person := range persons {
		t.Logf("{%v, %v}", person.Name, person.Age)
	}
}

type testPeople []testPerson

func (a testPeople) Len() int           { return len(a) }
func (a testPeople) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a testPeople) Less(i, j int) bool { return a[i].Age < a[j].Age }

func TestStructArraySort(t *testing.T) {
	people := []testPerson{
		{"Bob", 31},
		{"John", 42},
		{"Michael", 17},
		{"Jenny", 26},
	}

	t.Log("Origin map")
	for _, person := range people {
		t.Logf("{%v, %v}", person.Name, person.Age)
	}

	sort.Sort(testPeople(people))
	t.Log("After sort.Sort")
	for _, person := range people {
		t.Logf("{%v, %v}", person.Name, person.Age)
	}

	slices.SortFunc(people, func(a, b testPerson) int {
		return b.Age - a.Age
	})
	t.Log("After slices.SortFunc func(a, b Person) int { return b.Age - a.Age }")
	for _, person := range people {
		t.Logf("{%v, %v}", person.Name, person.Age)
	}
}
