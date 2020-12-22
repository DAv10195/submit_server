package util

// set of strings
type StringSet struct {
	Elements	map[string]bool	`json:"elements"`
}

// returns a new set of strings
func NewStringSet() *StringSet {
	return &StringSet{make(map[string]bool)}
}

// add the given elements to the set if not they are not present in it before
func (s *StringSet) Add(elements ...string) {
	for _, element := range elements {
		if _, found := s.Elements[element]; !found {
			s.Elements[element] = true
		}
	}
}

// remove the given elements from the set if they are not present in it
func (s *StringSet) Remove(elements ...string) {
	for element := range s.Elements {
		delete(s.Elements, element)
	}
}

// returns a boolean indicating if the given element is in the set or not
func (s *StringSet) Contains(element string) bool {
	if _, found := s.Elements[element]; found {
		return true
	}
	return false
}

// returns a slice with all elements of the set
func (s *StringSet) Slice() []string {
	var elements []string
	for element := range s.Elements {
		elements = append(elements, element)
	}
	return elements
}

// returns the number of elements in the set
func (s *StringSet) NumberOfElements() int {
	return len(s.Elements)
}
