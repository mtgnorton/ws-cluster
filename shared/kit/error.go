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
