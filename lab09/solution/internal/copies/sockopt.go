package copies

import "syscall"

func bindOpts(network, address string, c syscall.RawConn) error {
	var sockErr error
	if err := c.Control(func(fd uintptr) {
		for _, opt := range []int{syscall.SO_REUSEADDR, syscall.SO_REUSEPORT, syscall.SO_BROADCAST} {
			if e := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, opt, 1); e != nil {
				sockErr = e
				return
			}
		}
	}); err != nil {
		return err
	}
	return sockErr
}
