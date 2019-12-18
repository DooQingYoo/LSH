package main

import (
	"./utils"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

func main() {

	infochan := make(chan map[string]string)
	go utils.GetUserInfo(infochan)
	UserID, Z, pow := utils.GenerateZru()
	fmt.Println("用户ID为：")
	fmt.Println(UserID.String())
	g, err := utils.GetG()
	if err != nil {
		panic("can't get g: " + err.Error())
	}
	g1, err := utils.GetG1()
	if err != nil {
		panic("can't get g1: " + err.Error())
	}
	Ak, Bk, Ck := utils.GenerateABC(Z, pow, g, g1)
	fmt.Println("Ak = " + Ak.String())
	fmt.Println("Bk = " + Bk.String())
	fmt.Println("Ck = " + Ck.String())
	info := <-infochan
	hashAttr := utils.HashAttr(info, pow)
	fmt.Println("哈希结果为:")
	for _, v := range hashAttr {
		fmt.Println(v.String())
	}

	// 随机选择生成Ki，Li，Hxi的节点
	enNodeNumber := int(math.Ceil(float64(utils.NodeNumber) * 0.6))
	rand.Seed(time.Now().Unix())
	m := make(map[int]byte, enNodeNumber)
	for len(m) < enNodeNumber {
		if node := rand.Intn(utils.NodeNumber + 1); node != 0 {
			m[node] = 0x00
		}
	}
	selectedNodes := make([]int, enNodeNumber)
	index := 0
	for node := range m {
		selectedNodes[index] = node
		index++
	}
	sort.Ints(selectedNodes)
	before := time.Now()
	// 并行发送请求
	ch := make(chan bool, utils.NodeNumber)
	j := 0
	for i := 1; i <= utils.NodeNumber; i++ {
		if j < enNodeNumber && selectedNodes[j] == i {
			go utils.GetKLH(i, UserID, Ak, Bk, Ck, hashAttr, ch)
			j++
		} else {
			go utils.UselessFunc(i, ch)
		}
	}
	var ok bool
	for i := 0; i < utils.NodeNumber; i++ {
		select {
		case ok = <-ch:
			if !ok {
				fmt.Println("生成密钥失败")
				return
			}
		}
	}
	// 得到最终结果
	selectedNodesBytes := make([]byte, enNodeNumber)
	for k, v := range selectedNodes {
		selectedNodesBytes[k] = byte(v)
	}
	go utils.GetFinalKey(UserID, selectedNodesBytes, ch)
	for i := 2; i <= utils.NodeNumber; i++ {
		go utils.UselessFunc(i, ch)
	}
	for i := 0; i < utils.NodeNumber; i++ {
		select {
		case ok = <-ch:
			if !ok {
				fmt.Println("生成失败了")
				return
			}
		}
	}
	duration := time.Now().Sub(before)

	fmt.Println("##############################################")
	fmt.Println("耗时: " + duration.String())
}
