package session

import (
	"os"

	"github.com/virogg/networks-course/lab06/solution/pkg/ftp"
)

func (s *Session) HandleDELE(filename string) {
	path := s.RealPath(filename)
	if err := os.Remove(path); err != nil {
		s.Send(ftp.StatusActionNotTaken, "Delete failed")
		return
	}
	s.Send(ftp.StatusFileActionOK, "File deleted")
}
