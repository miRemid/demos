package mq

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func Test_read_close_channel(t *testing.T) {
	v := make(chan int)
	close(v)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	select {
	case vs, ok := <-v:
		t.Log(vs, ok)
	case <-ctx.Done():
		t.Log("failed")
	}
}

func Test_write_close_channel(t *testing.T) {
	v := make(chan int, 1)
	close(v)
	v <- 1
}

func Test_MQ(t *testing.T) {
	mq := NewMQ[int](WithBufferSize[int](5))
	for i := 0; i < 5; i++ {
		go func(i int) {
			pub, err := mq.NewPublisher("/key")
			if err != nil {
				panic(err)
			}
			cnt := 0
			for {
				time.Sleep(time.Microsecond * 10)
				value := rand.Intn(10)
				fmt.Printf("pub id = %d, value = %v\n", i, value)
				pub.Put(value)
				cnt++
				if cnt == 20 {
					break
				}
			}
		}(i)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := mq.NewSubscriber("/key", func(t int) {
				fmt.Printf("sub id = %d, value = %v\n", i, t)
			})
			if err != nil {
				panic(err)
			}
			time.Sleep((time.Duration(rand.Intn(10)) + 5) * time.Second)
		}(i)
	}
	wg.Wait()
	mq.Close()
}
