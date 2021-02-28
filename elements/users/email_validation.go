package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// validate the given email using the awesome Real Email API
func ValidateEmail(email string) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?email=%s", emailValidationUrl, email), nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var realEmailResp struct {
		Status string `json:"status"`
	}
	err = json.Unmarshal(body, &realEmailResp)
	if err != nil {
		return err
	}
	if realEmailResp.Status != valid {
		return &ErrEmailValidationFailed{email, realEmailResp.Status}
	}
	return nil
}
