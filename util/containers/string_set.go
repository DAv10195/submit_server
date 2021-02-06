package containers

import (
	"fmt"
	"strings"
	"sync"
)

// concurrent set of strings
type StringSet struct {
	Elements	map[string]int	`json:"elements"`
	mutex		*sync.RWMutex
}

// returns a new set of strings
func NewStringSet() *StringSet {
	return &StringSet{make(map[string]int), &sync.RWMutex{}}
}

// add the given elements to the set if not they are not present in it before
func (s *StringSet) Add(elements ...string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, element := range elements {
		if _, found := s.Elements[element]; !found {
			s.Elements[element] = 1
		}
	}
}

// remove the given elements from the set if they are present in it
func (s *StringSet) Remove(elements ...string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, element := range elements {
		if _, found := s.Elements[element]; found {
			delete(s.Elements, element)
		}
	}
}

// returns a boolean indicating if the given element is in the set or not
func (s *StringSet) Contains(element string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, found := s.Elements[element]
	return found
}

// returns a slice with all elements of the set
func (s *StringSet) Slice() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	var elements []string
	for element := range s.Elements {
		elements = append(elements, element)
	}
	return elements
}

// returns the number of elements in the set
func (s *StringSet) NumberOfElements() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.Elements)
}

func (s *StringSet) String() string {
	return fmt.Sprintf("{%s}", strings.Join(s.Slice(), ","))
}
