package portscan

import "fmt"

type Proto string

const (
	ProtoTCP  Proto = "tcp"
	ProtoUDP  Proto = "udp"
	ProtoBoth Proto = "both"
)

func ParseProto(s string) (Proto, error) {
	switch Proto(s) {
	case ProtoTCP, ProtoUDP, ProtoBoth:
		return Proto(s), nil
	default:
		return "", fmt.Errorf("invalid proto %q (want tcp|udp|both)", s)
	}
}

type Mode string

const (
	ModeAuto   Mode = "auto"
	ModeLocal  Mode = "local"
	ModeRemote Mode = "remote"
)

func ParseMode(s string) (Mode, error) {
	switch Mode(s) {
	case ModeAuto, ModeLocal, ModeRemote:
		return Mode(s), nil
	default:
		return "", fmt.Errorf("invalid mode %q (want auto|local|remote)", s)
	}
}
