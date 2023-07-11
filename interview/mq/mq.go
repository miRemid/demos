package mq

import (
	"context"
	"sync"
)

type Option[T any] func(*MQController[T])

func WithBufferSize[T any](size int) Option[T] {
	return func(m *MQController[T]) {
		m.bufferSize = size
	}
}

type MQ[T any] struct {
	_q    map[string]*MQController[T]
	mutex sync.Mutex

	opts []Option[T]
}

func NewMQ[T any](opts ...Option[T]) *MQ[T] {
	return &MQ[T]{
		_q:   make(map[string]*MQController[T]),
		opts: opts,
	}
}

func (mq *MQ[T]) NewPublisher(key string) (*Publisher[T], error) {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()
	_, ok := mq._q[key]
	if !ok {
		mq._q[key] = NewMQController[T](mq.opts...)
	}
	return mq._q[key].NewPub()
}

func (mq *MQ[T]) NewSubscriber(key string, callback func(T)) (*Subscriber[T], error) {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()
	_, ok := mq._q[key]
	if !ok {
		mq._q[key] = NewMQController[T](mq.opts...)
	}
	return mq._q[key].NewSub(callback)
}

func (mq *MQ[T]) Close() error {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()
	for k, v := range mq._q {
		v.Close()
		delete(mq._q, k)
	}
	return nil
}

type MQController[T any] struct {
	pub map[int]*Publisher[T]
	sub map[int]*Subscriber[T]
	v   chan T

	psize      int
	ssize      int
	bufferSize int
	rwmutex    sync.Mutex
}

func NewMQController[T any](opts ...Option[T]) *MQController[T] {
	mc := &MQController[T]{
		pub:        make(map[int]*Publisher[T], 0),
		sub:        make(map[int]*Subscriber[T], 0),
		bufferSize: 1024,
	}
	for _, o := range opts {
		o(mc)
	}
	mc.v = make(chan T, mc.bufferSize)
	return mc
}

func (mc *MQController[T]) NewPub() (*Publisher[T], error) {
	mc.rwmutex.Lock()
	defer mc.rwmutex.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	pub := &Publisher[T]{
		ctx:    ctx,
		cancel: cancel,
		id:     mc.psize,
		v:      mc.v,
		mc:     mc,
	}
	mc.pub[mc.psize] = pub
	mc.psize++
	return pub, nil
}

func (mc *MQController[T]) NewSub(callback func(T)) (*Subscriber[T], error) {
	mc.rwmutex.Lock()
	defer mc.rwmutex.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	sub := &Subscriber[T]{
		cancel:   cancel,
		ctx:      ctx,
		id:       mc.ssize,
		v:        mc.v,
		mc:       mc,
		callback: callback,
	}
	go sub.handle()
	mc.sub[mc.ssize] = sub
	mc.ssize++
	return sub, nil
}

func (mc *MQController[T]) Close() error {
	mc.rwmutex.Lock()
	defer mc.rwmutex.Unlock()
	close(mc.v)
	for k, v := range mc.pub {
		v.cancel()
		delete(mc.pub, k)
	}
	for k, v := range mc.sub {
		v.cancel()
		delete(mc.sub, k)
	}
	return nil
}

func (mc *MQController[T]) closeSub(id int) error {
	mc.rwmutex.Lock()
	defer mc.rwmutex.Unlock()
	mc.sub[id].cancel()
	delete(mc.sub, id)
	return nil
}

func (mc *MQController[T]) closePub(id int) error {
	mc.rwmutex.Lock()
	defer mc.rwmutex.Unlock()
	mc.pub[id].cancel()
	delete(mc.pub, id)
	return nil
}

type Publisher[T any] struct {
	ctx    context.Context
	cancel context.CancelFunc

	id int
	mc *MQController[T]
	v  chan T
}

func (pub *Publisher[T]) Put(v T) {
	select {
	case <-pub.ctx.Done():
		return
	default:
	}
	select {
	case <-pub.ctx.Done():
		return
	case pub.v <- v:
	}
}

func (pub *Publisher[T]) Close() error {
	return pub.mc.closePub(pub.id)
}

type Subscriber[T any] struct {
	ctx    context.Context
	cancel context.CancelFunc

	id       int
	mc       *MQController[T]
	v        chan T
	callback func(T)
}

func (sub *Subscriber[T]) handle() {
	for {
		select {
		case <-sub.ctx.Done():
			return
		default:
		}

		select {
		case v := <-sub.v:
			sub.callback(v)
		case <-sub.ctx.Done():
			return
		}
	}
}

func (sub *Subscriber[T]) Close() error {
	return sub.mc.closeSub(sub.id)
}
