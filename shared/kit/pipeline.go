package kit

type NextHandler func(any) any

// Handler 定义处理请求的接口
type Handler interface {
	Handle(request any, next NextHandler) any
}

type Pipeline interface {
	AddHandler(handler Handler) Pipeline
	Process(request any) any
}

// pipeline 定义管道结构
type pipeline struct {
	handlers []Handler
}

var DefaultPipeline = NewPipeline()

// NewPipeline 创建新的管道
func NewPipeline() Pipeline {
	return &pipeline{handlers: []Handler{}}
}

// AddHandler 添加处理器到管道
func (p *pipeline) AddHandler(handler Handler) Pipeline {
	p.handlers = append(p.handlers, handler)
	return p
}

// Process 处理请求
func (p *pipeline) Process(request any) any {
	handlers := make([]Handler, len(p.handlers))
	copy(handlers, p.handlers)
	var next NextHandler
	next = func(req any) any {
		// 如果没有更多的处理器，直接返回请求
		if len(handlers) == 0 {
			return req
		}
		if PipelineIsError(req) {
			return req
		}
		// 取出第一个处理器
		handler := handlers[0]
		handlers = handlers[1:]
		// 调用处理器
		return handler.Handle(req, next)
	}
	return next(request)
}

func PipelineAddHandler(handler Handler) Pipeline {
	return DefaultPipeline.AddHandler(handler)
}
func PipelineProcess(request any) any {
	return DefaultPipeline.Process(request)
}

func PipelineIsError(r any) bool {
	if _, ok := r.(error); ok {
		return true
	}
	return false
}
