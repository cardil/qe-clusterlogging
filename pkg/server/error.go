package server

type ErrorServer struct {
	Error error
}

func (e *ErrorServer) Run() error {
	return e.Error
}

func (e *ErrorServer) Kill() error {
	return nil
}

var _ Server = &ErrorServer{}
