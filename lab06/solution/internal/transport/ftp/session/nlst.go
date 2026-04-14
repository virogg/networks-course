package session

import (
	"fmt"
	"os"

	"github.com/virogg/networks-course/lab06/solution/pkg/ftp"
)

var fmtCRLF = "%s\r\n"

func (s *Session) HandleNLST() {
	dir := s.RealPath(s.Cwd)
	entries, err := os.ReadDir(dir)
	if err != nil {
		s.Send(ftp.StatusActionNotTaken, "Cannot list directory")
		return
	}

	dataConn, err := s.OpenDataConn()
	if err != nil {
		s.Send(ftp.StatusCantOpenDataConn, "Cannot open data connection")
		return
	}
	s.Send(ftp.StatusStartingDataTransfer, "Opening data connection")

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		fmt.Fprintf(dataConn, fmtCRLF, name)
	}
	dataConn.Close()
	s.Send(ftp.StatusOperationSuccessful, "Transfer complete")
}
