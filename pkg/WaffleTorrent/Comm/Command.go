package Comm

type CommandType uint8

const (
	CommandGet CommandType = iota
	CommandCancel
	CommandKill
)

type PeerCommand struct {
	Command CommandType
	Piece   int // signal which piece we wish to retrieve from peer
}

func GetCommand(piece int) *PeerCommand {
	return &PeerCommand{
		Command: CommandGet,
		Piece:   piece,
	}
}

func CancelCommand(piece int) *PeerCommand {
	return &PeerCommand{
		Command: CommandCancel,
		Piece:   piece,
	}
}

func KillCommand() *PeerCommand {
	return &PeerCommand{
		Command: CommandKill,
		Piece:   -1,
	}
}
