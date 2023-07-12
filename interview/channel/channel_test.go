package channel

import (
	"fmt"
	"sync"
	"testing"
)

func Test_printSeq(t *testing.T) {
	var chanA = make(chan struct{}, 1)
	var chanB = make(chan struct{}, 1)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				<-chanA
				fmt.Println(i)
				chanB <- struct{}{}
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			if i%2 != 0 {
				<-chanB
				fmt.Println(i)
				chanA <- struct{}{}
			}
		}
	}()
	chanA <- struct{}{}
	wg.Wait()
}

func Test_scan_file(t *testing.T) {
	/*
		协程池打印文件夹中所有文件名称
	*/

}
