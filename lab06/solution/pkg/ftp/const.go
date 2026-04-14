package ftp

const (
	StatusAlreadyOpen          = 125 // Data connection already open; transfer starting.
	StatusStartingDataTransfer = 150 // File status okay; about to open data connection.
	StatusOK                   = 200 // The requested action has been successfully completed.
	StatusSystemType           = 215 // NAME system type. Where NAME is an official system name from the registry kept by IANA.
	StatusServiceReady         = 220 // Service ready for new user.
	StatusServiceClosing       = 221 // Service closing control connection. Logged out if appropriate.
	StatusOperationSuccessful  = 226 // Closing data connection. Requested file action successful (for example, file transfer or file abort).
	StatusLoginSuccessful      = 230 // User logged in, proceed.
	StatusFileActionOK         = 250 // Requested file action was okay, completed.
	StatusCurrentDirectory     = 257
	StatusPasswordRequired     = 331 // Username okay, password needed.
	StatusCantOpenDataConn     = 425 // Can't open data connection.
	StatusTransferAborted      = 426 // Connection closed; transfer aborted.
	StatusSyntaxError          = 501 // Syntax error in parameters or arguments.
	StatusNotImplemented       = 502 // Command not implemented.
	StatusNotLoggedIn          = 530 // Not logged in.
	StatusActionNotTaken       = 550 // Requested action not taken. File unavailable (e.g., file not found, no access).
)
