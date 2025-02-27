package kit

import (
	"bytes"
	"fmt"
)

type baseError struct {
	filePosition string
	msg          any
	err          error
}

func (e *baseError) Error() string {

	buffer := bytes.NewBuffer([]byte{})
	// buffer.WriteString("{")

	buffer.WriteString(e.filePosition)
	if !IsNil(e.msg) {
		buffer.WriteString(",")
		buffer.WriteString(String(e.msg))
	}

	if e.err != nil {
		buffer.WriteString(" -> ")
		buffer.WriteString(e.err.Error())
	}

	return buffer.String()
}

func NewError(message any) error {
	if message == nil {
		return nil
	}
	return &baseError{
		msg:          message,
		filePosition: GetCallerPosition(2),
	}
}
func NewErrorf(format string, args ...any) error {
	return &baseError{
		msg:          fmt.Sprintf(format, args...),
		filePosition: GetCallerPosition(2),
	}
}

func TransmitError(err error) error {
	if err == nil {
		return nil
	}
	return &baseError{
		filePosition: GetCallerPosition(2),
		err:          err,
	}
}

func WrapError(err error, message any) error {
	if err == nil {
		return nil
	}
	return &baseError{
		msg:          message,
		filePosition: GetCallerPosition(2),
		err:          err,
	}
}
func WrapErrorf(err error, format string, args ...any) error {
	return &baseError{
		msg:          fmt.Sprintf(format, args...),
		filePosition: GetCallerPosition(2),
		err:          err,
	}
}

func CauseError(err error) error {

	for err != nil {
		if bErr, ok := err.(*baseError); ok && bErr.err != nil {
			err = bErr.err
		} else {
			break
		}
	}
	return err
}

// MultiError 合并多个错误, 如果oldErr为nil, 则返回errs中的第一个非nil错误, 否则返回oldErr和errs的组合
func MultiError(oldErr error, errs ...error) error {
	var nonNilErrs []error
	if oldErr != nil {
		nonNilErrs = append(nonNilErrs, oldErr)
	}
	for _, err := range errs {
		if err != nil {
			nonNilErrs = append(nonNilErrs, err)
		}
	}
	if len(nonNilErrs) == 0 {
		return nil
	}
	if len(nonNilErrs) == 1 {
		return nonNilErrs[0]
	}
	return &baseError{
		msg:          fmt.Sprintf("multiple errors: %v", nonNilErrs),
		filePosition: GetCallerPosition(2),
	}
}
