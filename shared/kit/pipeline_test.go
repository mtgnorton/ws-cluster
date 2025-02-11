package kit

import (
	"errors"
	"fmt"
	"testing"
)

type handle1 struct{}

func (m handle1) Handle(request any, next NextHandler) any {
	r := request.([]int)
	r = append(r, 2)
	if hr := next(r); PipelineIsError(hr) {
		return hr
	} else {
		r := hr.([]int)
		r = append(r, 3)
		return r
	}
}

type handler2 struct{}

func (m handler2) Handle(request any, next NextHandler) any {
	r := request.([]int)
	r = append(r, 4)

	if hr := next(r); PipelineIsError(hr) {
		return hr
	} else {
		r := hr.([]int)
		r = append(r, 5)
		return r
	}
}

type handler3 struct{}

func (m handler3) Handle(request any, next NextHandler) any {
	return errors.New("error")
}

type handler4 struct{}

func (m handler4) Handle(request any, next NextHandler) any {
	r := request.([]int)
	r = append(r, 6)
	if hr := next(r); PipelineIsError(hr) {
		return hr
	} else {
		r := hr.([]int)
		r = append(r, 7)
		return r
	}
}

func TestPipeline_Process(t *testing.T) {
	var s = []int{1}
	PipelineAddHandler(handle1{}).AddHandler(handler2{}).AddHandler(handler4{})
	result := PipelineProcess(s)
	if _, ok := result.([]int); !ok {
		t.Error("pipeline error")
		return
	}
	if fmt.Sprintf("%v", result) != "[1 2 4 6 7 5 3]" {
		t.Error("pipeline error")
	}
}

func TestPipeline_Process_IncludeError(t *testing.T) {
	var s = []int{1}
	PipelineAddHandler(handle1{}).AddHandler(handler2{}).AddHandler(handler3{}).AddHandler(handler4{})
	result := PipelineProcess(s)
	if !PipelineIsError(result) {
		t.Error("pipeline error")
		return
	}
}
