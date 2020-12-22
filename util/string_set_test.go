package util

import "testing"

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
