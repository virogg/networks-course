package session

import (
	"io"
	"os"

	"github.com/virogg/networks-course/lab06/solution/pkg/ftp"
)

func (s *Session) HandleRETR(filename string) {
	path := s.RealPath(filename)
	f, err := os.Open(path)
	if err != nil {
		s.Send(ftp.StatusActionNotTaken, "File not found")
		return
	}
	defer f.Close()

	dataConn, err := s.OpenDataConn()
	if err != nil {
		s.Send(ftp.StatusCantOpenDataConn, "Cannot open data connection")
		return
	}
	s.Send(ftp.StatusStartingDataTransfer, "Opening data connection")

	io.Copy(dataConn, f)
	dataConn.Close()
	s.Send(ftp.StatusOperationSuccessful, "Transfer complete")
}
