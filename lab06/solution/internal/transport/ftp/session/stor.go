package session

import (
	"io"
	"os"

	"github.com/virogg/networks-course/lab06/solution/pkg/ftp"
)

func (s *Session) HandleSTOR(filename string) {
	path := s.RealPath(filename)

	dataConn, err := s.OpenDataConn()
	if err != nil {
		s.Send(ftp.StatusCantOpenDataConn, "Cannot open data connection")
		return
	}
	defer dataConn.Close()

	s.Send(ftp.StatusStartingDataTransfer, "Opening data connection")

	data, err := io.ReadAll(dataConn)
	if err != nil {
		s.Send(ftp.StatusTransferAborted, "Transfer aborted")
		return
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		s.Send(ftp.StatusActionNotTaken, "Cannot store file")
		return
	}
	s.Send(ftp.StatusOperationSuccessful, "Transfer complete")
}
