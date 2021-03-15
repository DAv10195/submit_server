package containers

import (
	"fmt"
	"strings"
)

// set of strings
type StringSet struct {
	Elements	map[string]struct{}	`json:"elements"`
}

// returns a new set of strings
func NewStringSet() *StringSet {
	return &StringSet{make(map[string]struct{})}
}

// add the given elements to the set if not they are not present in it before
func (s *StringSet) Add(elements ...string) {
	for _, element := range elements {
		if _, found := s.Elements[element]; !found {
			s.Elements[element] = struct{}{}
		}
	}
}

// remove the given elements from the set if they are present in it
func (s *StringSet) Remove(elements ...string) {
	for _, element := range elements {
		if _, found := s.Elements[element]; found {
			delete(s.Elements, element)
		}
	}
}

// returns a boolean indicating if the given element is in the set or not
func (s *StringSet) Contains(element string) bool {
	_, found := s.Elements[element]
	return found
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

func (s *StringSet) String() string {
	return fmt.Sprintf("{%s}", strings.Join(s.Slice(), ","))
}

func StringSetUnion(sets ...*StringSet) *StringSet {
	unionSet := NewStringSet()
	for _, set := range sets {
		unionSet.Add(set.Slice()...)
	}
	return unionSet
}

func StringSetIntersection(sets ...*StringSet) *StringSet {
	intersectionSet := NewStringSet()
	if len(sets) > 0 {
		baseSet, restOfSets := sets[0], sets[1 :]
		for _, element := range baseSet.Slice() {
			foundInOtherSets := true
			for _, set := range restOfSets {
				foundInOtherSets = set.Contains(element)
			}
			if foundInOtherSets {
				intersectionSet.Add(element)
			}
		}
	}
	return intersectionSet
}
