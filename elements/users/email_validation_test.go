package users

import "testing"

func TestValidateEmail(t *testing.T) {
	validEmail := "david.abramov10@hotmail.com"
	err := ValidateEmail(validEmail)
	if err != nil {
		t.Fatalf("expected \"%s\" email to be valid but it is not. error: %v", validEmail, err)
	}
	invalidEmail := "invalid.david.abramov10@invalid.hotmail.com"
	err = ValidateEmail(invalidEmail)
	if err == nil {
		t.Fatalf("expected \"%s\" email to be invalid/unknown but it is valid", invalidEmail)
	}
}
