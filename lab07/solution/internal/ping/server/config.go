package server

type Option func(*Server)

func WithPort(port int) Option {
	return func(s *Server) {
		s.port = port
	}
}
