package berr

import "github.com/oldbai555/lbtool/pkg/lberr"

var (
	ErrToLargeMsgData     = lberr.NewErr(20001, "too large msg data receive")
	ErrNoPropertyFound    = lberr.NewErr(20002, "no property found")
	ErrConnectionClose    = lberr.NewErr(20003, "connection closed when send msg")
	ErrPackFail           = lberr.NewErr(20004, "Pack data is nil")
	ErrSendBuffMsgTimeOut = lberr.NewErr(20004, "send buff msg timeout")
)
