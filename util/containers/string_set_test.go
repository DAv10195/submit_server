package containers

import (
	"strings"
	"testing"
)

func TestStringSet_Add(t *testing.T) {
	set := NewStringSet()
	set.Add("azriel")
	if !set.Contains("azriel") {
		t.Fatal("\"azriel\" should be in the set but he's not")
	}
	set.Add("azriel")
	if set.NumberOfElements() != 1 {
		t.Fatalf("set has number of elements != 1")
	}

}

func TestStringSet_Remove(t *testing.T) {
	set := NewStringSet()
	set.Add("azriel")
	set.Remove("azriel")
	if set.Contains("azriel") {
		t.Fatalf("\"azriel\" is in the set although he shouldn't be")
	}
	if set.NumberOfElements() != 0 {
		t.Fatalf("set has number of elements != 0")
	}
}

func TestStringSet_Slice(t *testing.T) {
	david, nikita, azriel := "david", "nikita", "azriel"
	set := NewStringSet()
	set.Add(david, nikita, azriel)
	slice := set.Slice()
	if len(slice) != 3 {
		t.Fatalf("slice from set has number of elements != 3")
	}
	m := make(map[string]bool)
	m[david] = false
	m[nikita] = false
	m[azriel] = false
	for _, element := range slice {
		m[element] = true
	}
	for k, v := range m {
		if v != true {
			t.Fatalf("%s wasn't included in the slice generated from the set although it was in the set", k)
		}
	}
}

func TestStringSet_String(t *testing.T) {
	david, nikita, azriel := "david", "nikita", "azriel"
	set := NewStringSet()
	set.Add(david, nikita, azriel)
	setStr := set.String()
	if !strings.Contains(setStr, david) {
		t.Fatalf("%s doesn't contain %s", setStr, david)
	}
	if !strings.Contains(setStr, nikita) {
		t.Fatalf("%s doesn't contain %s", setStr, nikita)
	}
	if !strings.Contains(setStr, azriel) {
		t.Fatalf("%s doesn't contain %s", setStr, azriel)
	}
}

func TestUnion(t *testing.T) {
	a, b, c := "a", "b", "c"
	setA, setB, setC := NewStringSet(), NewStringSet(), NewStringSet()
	setA.Add(a)
	setB.Add(b)
	setC.Add(c)
	unionSet := StringSetUnion(setA, setB, setC)
	if !unionSet.Contains(a) || !unionSet.Contains(b) || !unionSet.Contains(c) {
		t.Fatalf("union set doesn't contain all elements from the given sets")
	}
}

func TestIntersection(t *testing.T) {
	a := "a"
	set1, set2, set3 := NewStringSet(), NewStringSet(), NewStringSet()
	set1.Add(a, "1")
	set2.Add(a, "2")
	set3.Add(a, "3")
	intersectionSet := StringSetIntersection(set1, set2, set3)
	if intersectionSet.NumberOfElements() != 1 {
		t.Fatal("intersection set contains no elements or more than 1 element")
	}
	if intersectionSet.Slice()[0] != a {
		t.Fatalf("intersection contains the wrong element")
	}
}
