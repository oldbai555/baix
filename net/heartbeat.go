package net

import (
	"fmt"
	"github.com/oldbai555/baix/iface"
	"github.com/oldbai555/lbtool/log"
	"time"
)

var _ iface.IHeartbeatChecker = (*HeartbeatChecker)(nil)

type HeartbeatChecker struct {
	interval time.Duration //  心跳检测时间间隔
	quitChan chan bool     // 退出信号

	makeMsg iface.HeartBeatMsgFunc // 自定义的心跳检测消息处理方法

	onRemoteNotAlive iface.OnRemoteNotAlive //  自定义的远程连接不存活时的处理方法

	msgID  uint32            // 心跳的消息ID
	router iface.IRouter     // 自定义的心跳检测消息业务处理路由
	conn   iface.IConnection // 绑定的链接

	beatFunc iface.HeartBeatFunc // 用户自定义心跳发送函数
}

func NewHeartbeatChecker(interval time.Duration) iface.IHeartbeatChecker {
	heartbeat := &HeartbeatChecker{
		interval: interval,
		quitChan: make(chan bool),

		// 使用默认的心跳消息生成函数和远程连接不存活时的处理方法
		makeMsg:          makeDefaultMsg,
		onRemoteNotAlive: notAliveDefaultFunc,
		msgID:            iface.HeartBeatDefaultMsgID,
		router:           &HeatBeatDefaultRouter{},
		beatFunc:         nil,
	}

	return heartbeat
}

func makeDefaultMsg(conn iface.IConnection) []byte {
	msg := fmt.Sprintf("heartbeat [%s->%s]", conn.LocalAddr(), conn.RemoteAddr())
	return []byte(msg)
}

func notAliveDefaultFunc(conn iface.IConnection) {
	log.Infof("Remote connection %s is not alive, stop it", conn.RemoteAddr())
	conn.Stop()
}

func (h *HeartbeatChecker) check() (err error) {

	if h.conn == nil {
		return nil
	}

	if !h.conn.IsAlive() {
		h.onRemoteNotAlive(h.conn)
		return nil
	}

	if h.beatFunc != nil {
		err = h.beatFunc(h.conn)
		if err != nil {
			log.Errorf("err is %v", err)
			return err
		}
		return nil
	}

	err = h.SendHeartBeatMsg()
	if err != nil {
		log.Errorf("err is %v", err)
		return err
	}

	return err
}

func (h *HeartbeatChecker) SetOnRemoteNotAliveFunc(alive iface.OnRemoteNotAlive) {
	h.onRemoteNotAlive = alive
}

func (h *HeartbeatChecker) SetHeartbeatMsgFunc(msgFunc iface.HeartBeatMsgFunc) {
	h.makeMsg = msgFunc
}

func (h *HeartbeatChecker) SetHeartbeatFunc(beatFunc iface.HeartBeatFunc) {
	h.beatFunc = beatFunc
}

func (h *HeartbeatChecker) BindRouter(msgID uint32, router iface.IRouter) {
	if router != nil && msgID != iface.HeartBeatDefaultMsgID {
		h.msgID = msgID
		h.router = router
	}
}

func (h *HeartbeatChecker) Start() {
	ticker := time.NewTicker(h.interval)
	for {
		select {
		case <-ticker.C:
			err := h.check()
			if err != nil {
				log.Errorf("err is %v", err)
			}
		case <-h.quitChan:
			ticker.Stop()
			return
		}
	}
}

func (h *HeartbeatChecker) Stop() {
	log.Infof("heartbeat checker stop, connID=%+v", h.conn.GetConnID())
	h.quitChan <- true
}

func (h *HeartbeatChecker) SendHeartBeatMsg() error {
	msg := h.makeMsg(h.conn)

	err := h.conn.SendMsg(h.msgID, msg)
	if err != nil {
		log.Errorf("send heartbeat msg error: %v, msgId=%+v msg=%+v", err, h.msgID, msg)
		return err
	}
	return nil
}

func (h *HeartbeatChecker) BindConn(conn iface.IConnection) {
	h.conn = conn
	conn.SetHeartBeat(h)
}

func (h *HeartbeatChecker) Clone() iface.IHeartbeatChecker {
	heartbeat := &HeartbeatChecker{
		interval:         h.interval,
		quitChan:         make(chan bool),
		beatFunc:         h.beatFunc,
		makeMsg:          h.makeMsg,
		onRemoteNotAlive: h.onRemoteNotAlive,
		msgID:            h.msgID,
		router:           h.router,
		conn:             nil, // The bound connection needs to be reassigned
	}

	return heartbeat
}

func (h *HeartbeatChecker) MsgID() uint32 {
	return h.msgID
}

func (h *HeartbeatChecker) Router() iface.IRouter {
	return h.router
}

type HeatBeatDefaultRouter struct {
	iface.BaseRouter
}

func (r *HeatBeatDefaultRouter) Handle(req iface.IRequest) {
	log.Infof("Recv Heartbeat from %s, MsgID = %+v, Data = %s", req.GetConnection().RemoteAddr(), req.GetMsgID(), string(req.GetData()))
}
