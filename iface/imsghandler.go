package iface

/*
Abstract layer of message management(消息管理抽象层)
*/

type IMsgHandle interface {
	// AddRouter 为消息添加具体的处理逻辑, msgID，支持整型，字符串
	AddRouter(msgID uint32, router IRouter)
	Use(Handlers ...RouterHandler)

	StartWorkerPool()                    //  Start the worker pool
	SendMsgToTaskQueue(request IRequest) // 将消息交给TaskQueue,由worker进行处理

	Execute(request IRequest) // Execute 执行责任链上的拦截器方法

	// AddInterceptor 注册责任链任务入口，每个拦截器处理完后，数据都会传递至下一个拦截器，使得消息可以层层处理层层传递，顺序取决于注册顺序
	AddInterceptor(interceptor IInterceptor)
}
