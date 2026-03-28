package bot

import "errors"

// 常见错误定义
var (
	ErrUnknownPacket = errors.New("unknown packet")
	ErrInvalidEcho   = errors.New("invalid echo")
	ErrInvalidArgs   = errors.New("invalid args")
	ErrInvalidRole   = errors.New("invalid role")
	ErrInvalidRounds = errors.New("invalid rounds")
)
