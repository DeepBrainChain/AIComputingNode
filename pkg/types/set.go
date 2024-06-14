package types

type None struct{}

var none None

type Set struct {
	elements map[interface{}]None
}

// func NewSet() *Set {
// 	return &Set{
// 		elements: make(map[interface{}]None),
// 	}
// }

func NewSet(items ...interface{}) *Set {
	s := &Set{
		elements: make(map[interface{}]None),
	}
	s.Add(items...)
	return s
}

// func (s *Set) Add(item interface{}) {
// 	s.elements[item] = none
// }

func (s *Set) Add(items ...interface{}) {
	for _, item := range items {
		s.elements[item] = none
	}
}

func (s *Set) Remove(item interface{}) {
	delete(s.elements, item)
}

func (s *Set) Elements() (result []interface{}) {
	for key := range s.elements {
		result = append(result, key)
	}
	return result
}

func (s *Set) Contains(item interface{}) bool {
	_, ok := s.elements[item]
	return ok
}

func (s *Set) Size() int {
	return len(s.elements)
}

func (s *Set) Clear() {
	s.elements = make(map[interface{}]None)
}
