package kit

import (
	"errors"
	"regexp"
	"strings"
	"testing"
)

func Test_NewError(t *testing.T) {

	t.Run("real_error_and_to_string", func(t *testing.T) {
		err := NewError("Test error")
		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		errStr := err.Error()
		if !strings.Contains(errStr, "Test error") {
			t.Errorf("Error message does not contain expected content, got: %s", errStr)
		}
		if matchFileAndLineAmount(errStr) != 1 {
			t.Errorf("Error message does not contain file location, got: %s", errStr)
		}
	})

	t.Run("nil_error", func(t *testing.T) {
		err := NewError(nil)
		if err != nil {
			t.Errorf("Expected nil error, but got: %v", err)
		}
	})

	t.Run("empty_str_error", func(t *testing.T) {
		err := NewError("")
		if err == nil {
			t.Errorf("Expected nil error, but got: %v", err)
		}
	})
}

func Test_TransmitError(t *testing.T) {
	t.Run("transmit_real_origin_error", func(t *testing.T) {
		err := errors.New("origin error")
		transmitErr := TransmitError(err)
		// kit/error_test.go:45 -> origin error
		if transmitErr == nil {
			t.Errorf("Expected an error, but got nil")
		}
		if matchFileAndLineAmount(transmitErr.Error()) != 1 {
			t.Errorf("Error message does not contain file location, got: %s", transmitErr.Error())
		}

		if !strings.HasSuffix(transmitErr.Error(), "origin error") {
			t.Errorf("Error message does not contain origin error, got: %s", transmitErr.Error())
		}

	})
	t.Run("transmit_real_kit_error", func(t *testing.T) {
		err := NewError("origin error")
		transmitErr := TransmitError(err)
		// kit/error_test.go:61 -> kit/error_test.go:60,origin error
		if transmitErr == nil {
			t.Errorf("Expected an error, but got nil")
		}
		if matchFileAndLineAmount(transmitErr.Error()) != 2 {
			t.Errorf("Error message does not contain file location, got: %s", transmitErr.Error())
		}
		if !strings.HasSuffix(transmitErr.Error(), "origin error") {
			t.Errorf("Error message does not contain origin error, got: %s", transmitErr.Error())
		}
	})
	t.Run("transmit_nil", func(t *testing.T) {
		transmitErr := TransmitError(nil)
		if transmitErr != nil {
			t.Error("Expected nil,but got error")
		}
	})
}

func Test_WrapError(t *testing.T) {
	t.Run("wrap_origin_error", func(t *testing.T) {
		err := errors.New("origin error")
		wrapErr := WrapError(err, "wrap error")
		// kit/error_test.go:84,wrap error -> origin error
		if wrapErr == nil {
			t.Errorf("Expected an error, but got nil")
		}
		if matchFileAndLineAmount(wrapErr.Error()) != 1 {
			t.Errorf("Error message does not contain file location, got: %s", wrapErr.Error())
		}
		if !strings.HasSuffix(wrapErr.Error(), "origin error") {
			t.Errorf("Error message does not contain origin error, got: %s", wrapErr.Error())
		}
		if !strings.Contains(wrapErr.Error(), "wrap error") {
			t.Errorf("Error message does not contain wrap error, got: %s", wrapErr.Error())
		}
	})
	t.Run("wrap_kit_error", func(t *testing.T) {
		err := NewError("origin error")
		wrapErr := WrapError(err, "wrap error")
		// kit/error_test.go:102,wrap error -> kit/error_test.go:101,origin error
		if wrapErr == nil {
			t.Errorf("Expected an error, but got nil")
		}
		if matchFileAndLineAmount(wrapErr.Error()) != 2 {
			t.Errorf("Error message does not contain file location, got: %s", wrapErr.Error())
		}
		if !strings.HasSuffix(wrapErr.Error(), "origin error") {
			t.Errorf("Error message does not contain origin error, got: %s", wrapErr.Error())
		}
		if !strings.Contains(wrapErr.Error(), "wrap error") {
			t.Errorf("Error message does not contain wrap error, got: %s", wrapErr.Error())
		}
	})

}

func Test_CauseError(t *testing.T) {
	t.Run("origin_error_is_self", func(t *testing.T) {
		err := errors.New("origin error")
		causeErr := CauseError(err)
		if causeErr.Error() != "origin error" {
			t.Errorf("Expected cause error to be nil, but got: %v", causeErr)
		}
	})
	t.Run("single_kit_error_is_self", func(t *testing.T) {
		err := NewError("kit error")
		causeErr := CauseError(err)
		if !isKitError(causeErr) {
			t.Errorf("Expected cause error to be kit error, but got: %v", causeErr)
		}
	})
	t.Run("kit_error_wrap_origin_error", func(t *testing.T) {
		err := WrapError(errors.New("origin error"), "wrap error")
		causeErr := CauseError(err)
		// origin error
		if causeErr == nil {
			t.Errorf("Expected cause error to be not nil, but got: %v", causeErr)
		}
		if causeErr.Error() != "origin error" {
			t.Errorf("Expected cause error to be origin error, but got: %v", causeErr)
		}
	})
	t.Run("kit_error_wrap_kit_error", func(t *testing.T) {
		err := WrapError(NewError("origin error"), "wrap error")
		causeErr := CauseError(err)
		// origin error
		if causeErr == nil {
			t.Errorf("Expected cause error to be not nil, but got: %v", causeErr)
		}
		if !isKitError(causeErr) {
			t.Errorf("Expected cause error to be kit error, but got: %v", causeErr)
		}
	})
}

// `_test\.go:\d+` 返回匹配次数
func matchFileAndLineAmount(errStr string) int {
	pattern := regexp.MustCompile(`_test\.go:\d+`)
	return len(pattern.FindAllStringIndex(errStr, -1))
}

func isKitError(err error) bool {
	_, ok := err.(*baseError)
	return ok
}
