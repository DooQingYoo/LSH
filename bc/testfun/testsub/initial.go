/*
 * @author DooQY
 * @createDate 2019/11/9 - 下午2:23
 */

package testsub

import "fmt"

var str string

func init() {
	str = "我是老杜"
	fmt.Println("init 方法被调用了")
}
func Mls() {
	fmt.Println("一个毫无卵用的方法")
	fmt.Println(str)
}
