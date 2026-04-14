package session

import (
	ftp2 "github.com/virogg/networks-course/lab06/solution/pkg/ftp"
)

func (s *Session) HandlePORT(arg string) {
	addr, err := ftp2.ParsePORTAddr(arg)
	if err != nil {
		s.Send(ftp2.StatusSyntaxError, "Invalid PORT")
		return
	}
	s.DataAddr = addr
	s.Send(ftp2.StatusOK, "PORT command successful")
}
