package main

import (
	"./utils"
	"fmt"
	"time"
)

func main() {
	var ok bool
	before := time.Now()
	ch := make(chan bool, utils.NodeNumber)
	for i := 1; i <= utils.NodeNumber; i++ {
		go utils.Dispatch(i, ch)
	}

	for i := 0; i < utils.NodeNumber; i++ {
		select {
		case ok = <-ch:
			if !ok {
				fmt.Println("初始化失败了")
				return
			}
		}
	}

	for i := 1; i <= utils.NodeNumber; i++ {
		go utils.Broadcast(i, ch)
	}

	for i := 0; i < utils.NodeNumber; i++ {
		select {
		case ok = <-ch:
			if !ok {
				fmt.Println("初始化失败了")
				return
			}
		}
	}

	go utils.Assemble(ch)
	for i := 2; i <= utils.NodeNumber; i++ {
		go utils.UselessFunc(i, ch)
	}

	for i := 0; i < utils.NodeNumber; i++ {
		select {
		case ok = <-ch:
			if !ok {
				fmt.Println("初始化失败了")
				return
			}
		}
	}

	cost := time.Now().Sub(before)
	fmt.Println("##############################################")
	fmt.Println("耗时：" + cost.String())
}
