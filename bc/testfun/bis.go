/*
 * @author DooQY
 * @createDate 2019/10/29 - 下午4:48
 */
package main

import (
	"fmt"

	"sync"
)

var num int64 = 0
var max = 10000
var wg sync.WaitGroup

func main() {
	wg.Add(2)
	go addNum()
	go addNum()
	wg.Wait()
	fmt.Printf("num=%d \n", num)
}
func addNum() {
	for i := 0; i < max; i++ {
		num++
	}
	wg.Done()
}
