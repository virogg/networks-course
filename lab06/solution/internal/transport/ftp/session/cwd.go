package session

import (
	"os"

	"github.com/virogg/networks-course/lab06/solution/pkg/ftp"
)

func (s *Session) HandleCWD(arg string) {
	newCwd := s.VirtualPath(arg)

	info, err := os.Stat(s.RealPath(newCwd))
	if err != nil || !info.IsDir() {
		s.Send(ftp.StatusActionNotTaken, "Directory not found")
		return
	}
	s.Cwd = newCwd
	s.Send(ftp.StatusFileActionOK, "Directory changed to "+s.Cwd)
}
