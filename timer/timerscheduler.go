package timer

/**
*  时间轮调度器
*  依赖模块，delayfunc.go  timer.go timewheel.go
 */

import (
	"context"
	"github.com/oldbai555/lbtool/log"
	"github.com/oldbai555/lbtool/pkg/routine"
	"math"
	"sync"
	"time"
)

const (
	//MaxChanBuff 默认缓冲触发函数队列大小
	MaxChanBuff = 2048
	//MaxTimeDelay 默认最大误差时间
	MaxTimeDelay = 100
)

// Scheduler 计时器调度器
type Scheduler struct {
	//当前调度器的最高级时间轮
	tw *TimeWheel
	//定时器编号累加器
	IDGen uint32
	//已经触发定时器的channel
	triggerChan chan *DelayFunc
	//互斥锁
	sync.RWMutex
}

// NewScheduler 返回一个定时器调度器 ，主要创建分层定时器，并做关联，并依次启动
func NewScheduler() *Scheduler {

	//创建秒级时间轮
	secondTw := NewTimeWheel(SecondName, SecondInterval, SecondScales, TimersMaxCap)
	//创建分钟级时间轮
	minuteTw := NewTimeWheel(MinuteName, MinuteInterval, MinuteScales, TimersMaxCap)
	//创建小时级时间轮
	hourTw := NewTimeWheel(HourName, HourInterval, HourScales, TimersMaxCap)

	//将分层时间轮做关联
	hourTw.AddTimeWheel(minuteTw)
	minuteTw.AddTimeWheel(secondTw)

	//时间轮运行
	secondTw.Run()
	minuteTw.Run()
	hourTw.Run()

	return &Scheduler{
		tw:          hourTw,
		triggerChan: make(chan *DelayFunc, MaxChanBuff),
	}
}

// CreateTimerAt 创建一个定点Timer 并将Timer添加到分层时间轮中， 返回Timer的tID
func (ts *Scheduler) CreateTimerAt(df *DelayFunc, unixNano int64) (uint32, error) {
	ts.Lock()
	defer ts.Unlock()

	ts.IDGen++
	return ts.IDGen, ts.tw.AddTimer(ts.IDGen, NewTimerAt(df, unixNano))
}

// CreateTimerAfter 创建一个延迟Timer 并将Timer添加到分层时间轮中， 返回Timer的tID
func (ts *Scheduler) CreateTimerAfter(df *DelayFunc, duration time.Duration) (uint32, error) {
	ts.Lock()
	defer ts.Unlock()

	ts.IDGen++
	return ts.IDGen, ts.tw.AddTimer(ts.IDGen, NewTimerAfter(df, duration))
}

// CancelTimer 删除timer
func (ts *Scheduler) CancelTimer(tID uint32) {
	ts.Lock()
	defer ts.Unlock()

	tw := ts.tw
	for tw != nil {
		tw.RemoveTimer(tID)
		tw = tw.nextTimeWheel
	}
}

// GetTriggerChan 获取计时结束的延迟执行函数通道
func (ts *Scheduler) GetTriggerChan() chan *DelayFunc {
	return ts.triggerChan
}

// Start 非阻塞的方式启动timerSchedule
func (ts *Scheduler) Start() {
	routine.Go(context.TODO(), func(ctx context.Context) error {
		for {
			//当前时间
			now := UnixMilli()
			//获取最近MaxTimeDelay 毫秒的超时定时器集合
			timerList := ts.tw.GetTimerWithIn(MaxTimeDelay * time.Millisecond)
			for _, timer := range timerList {
				if math.Abs(float64(now-timer.unixts)) > MaxTimeDelay {
					//已经超时的定时器，报警
					log.Warnf("want call at %d; real call at %d ; delay %d", timer.unixts, now, now-timer.unixts)
				}
				ts.triggerChan <- timer.delayFunc
			}
			time.Sleep(MaxTimeDelay / 2 * time.Millisecond)
		}
	})
}

// NewAutoExecScheduler 时间轮定时器 自动调度
func NewAutoExecScheduler() *Scheduler {
	//创建一个调度器
	autoExecScheduler := NewScheduler()
	//启动调度器
	autoExecScheduler.Start()

	//永久从调度器中获取超时 触发的函数 并执行
	routine.Go(context.TODO(), func(ctx context.Context) error {
		delayFuncChan := autoExecScheduler.GetTriggerChan()
		for df := range delayFuncChan {
			asyncDoDelayFunc(ctx, df)
		}
		return nil
	})

	return autoExecScheduler
}

func asyncDoDelayFunc(ctx context.Context, df *DelayFunc) {
	routine.Go(ctx, func(ctx context.Context) error {
		df.Call()
		return nil
	})
}
