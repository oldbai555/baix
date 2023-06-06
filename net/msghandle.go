package net

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/oldbai555/baix/conf"
	"github.com/oldbai555/baix/iface"
	"github.com/oldbai555/lbtool/log"
	"github.com/oldbai555/lbtool/pkg/routine"
	"sync"
)

const (
	// WorkerIDWithoutWorkerPool 如果不启动Worker协程池，则会给MsgHandler分配一个虚拟的 WorkerID，这个workerID为0, 便于指标统计,启动了Worker协程池后，每个worker的ID为0,1,2,3...
	WorkerIDWithoutWorkerPool int = 0
)

var _ iface.IMsgHandle = (*MsgHandle)(nil)
var _ iface.IInterceptor = (*MsgHandle)(nil)

type MsgHandle struct {
	// 存放每个MsgID 所对应的处理方法的map属性
	Apis map[uint32]iface.IRouter

	// 业务工作Worker池的数量
	WorkerPoolSize uint32

	// Worker负责取任务的消息队列
	TaskQueue []chan iface.IRequest

	// builder 责任链构造器
	builder *chainBuilder

	apiLock *sync.RWMutex // 并发互斥
}

func newMsgHandle() *MsgHandle {
	handle := &MsgHandle{
		Apis:           make(map[uint32]iface.IRouter),
		WorkerPoolSize: conf.GlobalConfig.WorkerPoolSize,
		TaskQueue:      make([]chan iface.IRequest, conf.GlobalConfig.WorkerPoolSize), // 一个worker对应一个queue
		builder:        newChainBuilder(),
		apiLock:        new(sync.RWMutex),
	}

	// 此处必须把 msgHandler 添加到责任链中，并且是责任链最后一环，在 msgHandler 中进行解码后由 router 做数据分发
	handle.builder.Tail(handle)
	return handle
}

// StartOneWorker 启动一个Worker工作流程
func (m *MsgHandle) StartOneWorker(workerID int, taskQueue chan iface.IRequest) {
	log.Infof("【BaiX】Worker ID = %d is started.", workerID)
	// 不断地等待队列中的消息
	for {
		select {
		// 有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			switch req := request.(type) {
			case iface.IFuncRequest:
				// 内部函数调用request
				m.doFuncHandler(req, workerID)
			case iface.IRequest:
				m.doMsgHandler(req, workerID)
			}
		}
	}
}

// doFuncHandler 执行函数式请求
func (m *MsgHandle) doFuncHandler(request iface.IFuncRequest, workerId int) {
	// 执行函数式请求
	request.CallFunc()
}

// doMsgHandler 立即以非阻塞方式处理消息)
func (m *MsgHandle) doMsgHandler(request iface.IRequest, workerID int) {
	msgId := request.GetMsgID()
	handler, ok := m.Apis[msgId]
	if !ok {
		log.Errorf("【BaiX】api msgID = %d is not FOUND!", request.GetMsgID())
		return
	}

	// Request 请求绑定Router对应关系
	request.BindRouter(handler)

	// Execute the corresponding processing method
	request.Call()
}

func (m *MsgHandle) GetTaskQueueWorkerId(request iface.IRequest) uint64 {
	// 根据ConnID来分配当前的连接应该由哪个worker负责处理
	// 轮询的平均分配法则
	// 得到需要处理此条连接的workerID
	workerID := request.GetConnection().GetConnID() % uint64(m.WorkerPoolSize)
	return workerID
}

func (m *MsgHandle) Intercept(chain iface.IChain) iface.IcResp {
	request := chain.Request()
	if request != nil {
		switch request.(type) {
		case iface.IRequest:
			iRequest := request.(iface.IRequest)
			if conf.GlobalConfig.WorkerPoolSize > 0 {
				// 已经启动工作池机制，将消息交给Worker处理
				m.SendMsgToTaskQueue(iRequest)
			} else {
				// 从绑定好的消息和对应的处理方法中执行对应的 Handle 方法
				routine.Go(iRequest.GetConnection().Context(), func(ctx context.Context) error {
					m.doMsgHandler(iRequest, WorkerIDWithoutWorkerPool)
					return nil
				})
			}
		}
	}

	return chain.Proceed(chain.Request())
}

func (m *MsgHandle) AddRouter(msgID uint32, router iface.IRouter) {
	m.apiLock.Lock()
	defer m.apiLock.Unlock()

	// 1. 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := m.Apis[msgID]; ok {
		msgErr := fmt.Sprintf("repeated api , msgID = %+v\n", msgID)
		panic(msgErr)
	}

	// 2. 添加msg与api的绑定关系
	m.Apis[msgID] = router
	log.Infof("Add Router msgID = %d", msgID)
}

func (m *MsgHandle) StartWorkerPool() {
	// Iterate through the required number of workers and start them one by one
	// (遍历需要启动worker的数量，依此启动)
	for i := 0; i < int(m.WorkerPoolSize); i++ {
		// A worker is started
		// Allocate space for the corresponding task queue for the current worker
		// (给当前worker对应的任务队列开辟空间)
		var idx = i
		m.TaskQueue[idx] = make(chan iface.IRequest, conf.GlobalConfig.MaxWorkerTaskLen)

		// Start the current worker, blocking and waiting for messages to be passed in the corresponding task queue
		// (启动当前Worker，阻塞的等待对应的任务队列是否有消息传递进来)
		routine.Go(context.TODO(), func(ctx context.Context) error {
			m.StartOneWorker(idx, m.TaskQueue[idx])
			return nil
		})
	}
}

func (m *MsgHandle) SendMsgToTaskQueue(request iface.IRequest) {
	workerID := m.GetTaskQueueWorkerId(request)
	m.TaskQueue[workerID] <- request
	log.Debugf("【BaiX】SendMsgToTaskQueue --> %s", hex.EncodeToString(request.GetData()))
}

func (m *MsgHandle) Execute(request iface.IRequest) {
	// 将消息丢到责任链，通过责任链里拦截器层层处理层层传递
	m.builder.Execute(request)
}

func (m *MsgHandle) AddInterceptor(interceptor iface.IInterceptor) {
	if m.builder != nil {
		m.builder.AddInterceptor(interceptor)
	}
}
