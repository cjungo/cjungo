package cjungo

import (
	"encoding/json"
	"fmt"
)

type ApiError struct {
	Code     int   `json:"code"`
	Message  any   `json:"message"`
	HttpCode int   `json:"-"`
	Reason   error `json:"-"`
}

func (err *ApiError) Error() string {
	if result, err := json.Marshal(err); err != nil {
		return fmt.Sprintf("ApiError JSON marshal failed: %v", err)
	} else {
		return string(result)
	}
}
