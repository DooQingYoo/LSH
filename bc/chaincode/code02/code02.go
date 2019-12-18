/*
 * @author DooQY
 * @createDate 2019/10/28 - 下午2:11
 */

package main

import (
	"fmt"
	"github.com/Nik-U/pbc"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"math"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// parameters of the curve
const str = `type a
	q 8780710799663312522437781984754049815806883199414208211028653399266475630880222957078625179422662221423155858769582317459277713367317481324925129998224791
	h 12016012264891146079388821366740534204802954401251311822919615131047207289359704531102844802183906537786776
	r 730750818665451621361119245571504901405976559617
	exp2 159
	exp1 107
	sign1 1
	sign0 1`

var pairing, _ = pbc.NewPairingFromString(str)

//

type FirstTemp struct {
}

/**
生成 g，g1；
把一共有多少个节点和 g，g1 一起保存起来
*/
// 应该传入的参数：节点的个数，int类型
func (f FirstTemp) Init(stub shim.ChaincodeStubInterface) peer.Response {

	// 存储节点个数
	args := stub.GetStringArgs()
	if len(args) != 1 {
		return shim.Error("Please input the number of the nodes")
	}
	number, e := strconv.Atoi(args[0])
	if e != nil {
		return shim.Error("Please input the number of the nodes")
	}
	err := stub.PutState("NodeNumber", []byte{byte(number)})
	if err != nil {
		return shim.Error("Can't put node number: " + err.Error())
	}

	// 计算g，g1，e(g,g)
	g := pairing.NewG1().Rand()
	a := pairing.NewZr().Rand()
	g1 := pairing.NewG1().PowZn(g, a)
	es := pairing.NewGT()
	es.Pair(g, g)
	Gbytes := g.Bytes()
	G1bytes := g1.Bytes()
	Ebytes := es.Bytes()

	err = stub.PutState("g", Gbytes)
	if err != nil {
		return shim.Error("Can't put g: " + err.Error())
	}
	err = stub.PutState("g1", G1bytes)
	if err != nil {
		return shim.Error("Can't put g1: " + err.Error())
	}
	err = stub.PutState("e", Ebytes)
	if err != nil {
		return shim.Error("Can't put e: " + err.Error())
	}

	suc := "g = " + g.String() + "ng1 = " + g1.String()
	return shim.Success([]byte(suc))
}

func (f FirstTemp) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fun, _ := stub.GetFunctionAndParameters()
	switch fun {
	case "init":
		return compile()
	case "dispatch":
		return dispatch(stub)
	case "queryBytes":
		return queryBytes(stub)
	case "broadcast":
		return broadcast(stub)
	case "assemble":
		return assemble(stub)
	case "getKAndL":
		return getKAndL(stub)
	case "getFinalKey":
		return getFinalKey(stub)
	case "useless":
		return useless(stub)
	}
	return shim.Error("No such function!")
}

func compile() peer.Response {
	return shim.Success([]byte("Compile finished"))
}

// 给SDK专门准备的，不再需要先转成string了
func queryBytes(stub shim.ChaincodeStubInterface) peer.Response {

	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Please input what you want to query!")
	}
	key := args[0]
	bytes2, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Wrong when GetState(): " + err.Error())
	}
	return shim.Success(bytes2)
}

func dispatch(stub shim.ChaincodeStubInterface) peer.Response {

	// 找出本节点的序号和一共有多少个节点
	nID := getNodeID(stub)
	nid, err := strconv.Atoi(nID)
	if err != nil {
		return shim.Error("Can't convert node ID to integer: " + err.Error())
	}
	n, err := stub.GetState("NodeNumber")
	if err != nil {
		return shim.Error("Can't get node number: " + err.Error())
	}
	nodeNumber := int(n[0])

	// 生成这个节点的所有的 sij
	enNodeNumber := int(math.Ceil(float64(nodeNumber) * 0.6))
	si := pairing.NewZr().Rand()
	cs := GenCoef(enNodeNumber)
	sij := make([]*pbc.Element, nodeNumber)
	for j := 1; j <= nodeNumber; j++ {
		sij[j-1] = C(cs, j, si)
	}

	// 把这个节点的 sij 保存到链上，sii除外
	var ret string
	for j := 1; j <= nodeNumber; j++ {

		if j == nid {
			continue
		}

		key, err := stub.CreateCompositeKey("sij", []string{nID, strconv.Itoa(j)})
		if err != nil {
			return shim.Error("Can't create composite key: " + err.Error())
		}
		err = stub.PutState(key, sij[j-1].Bytes())
		if err != nil {
			return shim.Error("wrong with the putState: " + err.Error())
		}

		ret = ret + "[sij" + nID + strconv.Itoa(j) + ": " + sij[j-1].String() + "]  "
	}

	// 保存 sii 为私有数据
	collection := "collection" + nID
	err = stub.PutPrivateData(collection, "sii"+nID, sij[nid-1].Bytes())
	if err != nil {
		return shim.Error("Can't save sii: " + err.Error())
	}

	return shim.Success([]byte(ret))
}

func broadcast(stub shim.ChaincodeStubInterface) peer.Response {

	// 找出本节点的序号和一共有多少个节点
	nID := getNodeID(stub)
	nid, err := strconv.Atoi(nID)
	if err != nil {
		return shim.Error("Can't convert node ID to integer: " + nID)
	}
	n, err := stub.GetState("NodeNumber")
	if err != nil {
		return shim.Error("Can't get node number: " + err.Error())
	}
	nodeNumber := int(n[0])

	// 收集所需要的 sij
	sij := make([]*pbc.Element, nodeNumber)
	for i := 1; i <= nodeNumber; i++ {
		if i == nid {
			continue
		}
		key, err := stub.CreateCompositeKey("sij", []string{strconv.Itoa(i), nID})
		if err != nil {
			return shim.Error("Can't create composite key: " + err.Error())
		}
		// 要一直等到所有的sij都能取到为止
		var bytes2 []byte = nil
		for bytes2 == nil {
			bytes2, err = stub.GetState(key)
			if err != nil {
				return shim.Error("Wrong when GetState(): " + err.Error())
			}
		}
		sij[i-1] = pairing.NewZr().SetBytes(bytes2)
	}

	// 计算ski，求和，存储私有数据
	collection := "collection" + nID
	bytes2, err := stub.GetPrivateData(collection, "sii"+nID)
	if err != nil {
		return shim.Error("Can't get sii" + nID + ": " + err.Error())
	}
	if bytes2 == nil {
		return shim.Error("The node has no sii !")
	}
	ski := pairing.NewZr().SetBytes(bytes2)
	for i := 0; i < nodeNumber; i++ {
		if i+1 == nid {
			continue
		}
		ski.Add(ski, sij[i])
	}
	err = stub.PutPrivateData(collection, "ski"+nID, ski.Bytes())
	if err != nil {
		return shim.Error("Can't save private data ski: " + err.Error())
	}

	// 计算pki，存入账本
	Ebytes, err := stub.GetState("e")
	if err != nil {
		return shim.Error("Can't get e: " + err.Error())
	}
	es := pairing.NewGT().SetBytes(Ebytes)
	es.PowZn(es, ski)
	Ebytes = es.Bytes()
	err = stub.PutState("pki"+nID, Ebytes)
	if err != nil {
		return shim.Error("Can't put pki to state: " + err.Error())
	}
	return shim.Success(Ebytes)
}

func assemble(stub shim.ChaincodeStubInterface) peer.Response {

	nID := getNodeID(stub)
	nid, err := strconv.Atoi(nID)
	if err != nil {
		return shim.Error("Can't convert node ID to integer: " + nID)
	}
	n, err := stub.GetState("NodeNumber")
	if err != nil {
		return shim.Error("Can't get node number: " + err.Error())
	}
	nodeNumber := int(n[0])
	enNodeNumber := int(math.Ceil(float64(nodeNumber) * 0.6))

	// 随机选择t-1个用于计算的节点
	rand.Seed(time.Now().Unix())
	m := make(map[int]byte, enNodeNumber-1)
	for len(m) < enNodeNumber-1 {
		if node := rand.Intn(nodeNumber + 1); node != nid && node != 0 {
			m[node] = 0x00
		}
	}
	selectedNodes := make([]int, enNodeNumber)
	selectedNodes[0] = nid
	index := 1
	for node := range m {
		selectedNodes[index] = node
		index++
	}
	sort.Ints(selectedNodes)

	// 按照选出的节点号把相应的pki取出来
	S := make([]*pbc.Element, enNodeNumber)
	for i := 0; i < enNodeNumber; i++ {
		node := selectedNodes[i]

		// 等到在账本里出现该节点的pki为止
		var bytes2 []byte = nil
		for bytes2 == nil {
			bytes2, err = stub.GetState("pki" + strconv.Itoa(node))
			if err != nil {
				return shim.Error("Can't get pki" + strconv.Itoa(node) + ": " + err.Error())
			}
		}
		S[i] = pairing.NewGT().SetBytes(bytes2)
	}

	// 计算e(g,g)alpha，应该是每个节点的结果都一样的才对
	Xl := make([]*pbc.Element, enNodeNumber)
	for i := 0; i < enNodeNumber; i++ {
		Xl[i] = pairing.NewZr().SetInt32(int32(selectedNodes[i]))
	}
	Pkis := make([]*pbc.Element, enNodeNumber)
	for i := 0; i < enNodeNumber; i++ {
		l := L(Xl, i)
		Pkis[i] = pairing.NewGT().PowZn(S[i], l)
	}
	publicE := PIgt(Pkis)

	err = stub.PutState("publicE", publicE.Bytes())
	if err != nil {
		return shim.Error("Can't put publicE: " + err.Error())
	}
	return shim.Success(publicE.Bytes())
}

// 传入的参数：
// 1. UserID，[]byte
// 2. Ak, []byte
// 3. Bk, []byte
// 4. Ck, []byte
// 5. Hx, [][]byte
func getKAndL(stub shim.ChaincodeStubInterface) peer.Response {
	// 拿到所有传入的参数
	args := stub.GetArgs()
	if len(args) < 7 {
		return shim.Error("Short of arguments")
	}
	uID := pairing.NewZr().SetBytes(args[1])
	Ak := pairing.NewG1().SetBytes(args[2])
	Bk := pairing.NewG1().SetBytes(args[3])
	Ck := pairing.NewG1().SetBytes(args[4])
	Hx := make([]*pbc.Element, len(args)-5)
	for i := 5; i < len(args); i++ {
		Hx[i-5] = pairing.NewG1().SetBytes(args[i])
	}

	// 拿到自己节点的ski, 随机数bi
	nID := getNodeID(stub)
	data2, err := stub.GetPrivateData("collection"+nID, "ski"+nID)
	if err != nil {
		return shim.Error("Can't get ski of node" + nID + ": " + err.Error())
	}
	if data2 == nil {
		return shim.Error("The node" + nID + " has no ski")
	}
	ski := pairing.NewZr().SetBytes(data2)
	bi := pairing.NewZr().Rand()

	// 计算 Ki, Li, Kattri
	Ki := pairing.NewG1().Set1()
	Li := pairing.NewG1().Set1()
	Ak.PowZn(Ak, ski)
	Bk.PowZn(Bk, bi)
	Ki.Mul(Ak, Bk)
	Li.PowZn(Ck, bi)
	for _, v := range Hx {
		v.PowZn(v, bi)
	}

	// 公开（存储）本节点的计算结果
	key, err := stub.CreateCompositeKey("KLK", []string{uID.String(), nID})
	if err != nil {
		return shim.Error("Can't create composite key: " + err.Error())
	}
	var value []string
	value = append(value, Ki.String())
	value = append(value, Li.String())
	for _, v := range Hx {
		value = append(value, v.String())
	}
	join := strings.Join(value, ";")
	err = stub.PutState(key, []byte(join))
	if err != nil {
		return shim.Error("Can't put state: " + err.Error())
	}
	return shim.Success([]byte(join))
}

// 传入的参数：
// 1. UserID, []byte
// 2. selectedNodes []byte
func getFinalKey(stub shim.ChaincodeStubInterface) peer.Response {

	// 准备：得到节点数和参与计算的节点数，初始化切片
	args := stub.GetArgs()
	if len(args) < 3 {
		return shim.Error("The UserID and selected nodes should be input")
	}
	UserID := pairing.NewZr().SetBytes(args[1])
	selectedNodes := args[2]
	enNodeNumber := len(selectedNodes)
	Kis := make([]*pbc.Element, enNodeNumber)
	Lis := make([]*pbc.Element, enNodeNumber)
	His := make([][]*pbc.Element, enNodeNumber)

	// 从账本中取出K，L，H的全部的值
	for i := 0; i < enNodeNumber; i++ {
		nID := int(selectedNodes[i])
		key, err := stub.CreateCompositeKey("KLK", []string{UserID.String(), strconv.Itoa(nID)})
		if err != nil {
			return shim.Error("Can't create composite key: " + err.Error())
		}
		var state []byte = nil
		for state == nil {
			state, err = stub.GetState(key)
			if err != nil {
				return shim.Error("Can't get KLK from node" + strconv.Itoa(nID) + " : " + err.Error())
			}
		}
		str := strings.Split(string(state), ";")
		temp, ok := pairing.NewG1().SetString(str[0], 10)
		if !ok {
			return shim.Error("Can't convert K of node" + strconv.Itoa(nID))
		}
		Kis[i] = temp
		temp, ok = pairing.NewG1().SetString(str[1], 10)
		if !ok {
			return shim.Error("Can't convert L of node" + strconv.Itoa(nID))
		}
		Lis[i] = temp
		for j := 2; j < len(str); j++ {
			temp, ok = pairing.NewG1().SetString(str[j], 10)
			His[i] = append(His[i], temp)
		}
	}

	// 计算最后的K，L，H
	Xl := make([]*pbc.Element, enNodeNumber)
	for i := 0; i < enNodeNumber; i++ {
		Xl[i] = pairing.NewZr().SetInt32(int32(selectedNodes[i]))
	}
	for i := 0; i < enNodeNumber; i++ {
		l := L(Xl, i)
		Kis[i].PowZn(Kis[i], l)
		Lis[i].PowZn(Lis[i], l)
		for j := 0; j < len(His[i]); j++ {
			His[i][j].PowZn(His[i][j], l)
		}
	}
	finalK := PIg1(Kis)
	finalL := PIg1(Lis)
	finalH := make([]*pbc.Element, len(His[0]))
	for i := 0; i < len(His[0]); i++ {
		accm := pairing.NewG1().Set1()
		for j := 0; j < len(His); j++ {
			accm.Mul(accm, His[j][i])
		}
		finalH[i] = accm
	}
	// K，L，H分别存入账本，存入的键为 UserID+K，L，H，只有H是先转换为string再组合存起来，K和L是直接存字节
	uID := UserID.String()
	err := stub.PutState(uID+"K", finalK.Bytes())
	if err != nil {
		return shim.Error("Can't put K: " + err.Error())
	}
	err = stub.PutState(uID+"L", finalL.Bytes())
	if err != nil {
		return shim.Error("Can't put L: " + err.Error())
	}
	var Hs []string
	for _, v := range finalH {
		Hs = append(Hs, v.String())
	}
	H := strings.Join(Hs, ";")
	err = stub.PutState(uID+"H", []byte(H))
	if err != nil {
		return shim.Error("Can't put H: " + err.Error())
	}
	ret := finalK.String() + ";" + finalL.String() + ";" + H

	return shim.Success([]byte(ret))
}

func getNodeID(stub shim.ChaincodeStubInterface) string {

	bytes2, err := stub.GetCreator()
	if err != nil {
		return "Can't find creator: " + err.Error()
	}

	cert := string(bytes2)
	index := strings.Index(cert, "MSP")

	orgInfo := string(bytes2[1:index])
	reg, err := regexp.Compile(`\d+`)
	if err != nil {
		return "Can't create reg: " + err.Error()
	}
	org := reg.FindString(orgInfo)
	return org
}

// C returns the result of f(x) = si+C1*x+C2*x^2+...+Ctk-1*x^(tk-1)
func C(Cs []*pbc.Element, index int, si *pbc.Element) *pbc.Element {
	x := pairing.NewZr().SetInt32(int32(index))
	ret := pairing.NewZr().Set0()
	for _, coef := range Cs {
		ret.Add(ret, coef)
		ret.MulZn(ret, x)
	}
	return ret.Add(ret, si)
}

// GenCoef generates the coefficients used by func C
func GenCoef(t int) []*pbc.Element {
	Cs := make([]*pbc.Element, t-1)
	for i := 0; i < t-1; i++ {
		Cs[i] = pairing.NewZr().Rand()
	}
	return Cs
}

// L generates the Lagrange coefficient of an index
func L(SelectedNodes []*pbc.Element, indexOfArray int) *pbc.Element { // indexOfArray represents index of SelectedNodes
	t := len(SelectedNodes)
	others := make([]*pbc.Element, t)
	copy(others, SelectedNodes)
	cur := others[indexOfArray]
	if indexOfArray == t-1 {
		others = others[:indexOfArray]
	} else {
		others = append(others[:indexOfArray], others[indexOfArray+1:]...)
	}
	tmpN := make([]*pbc.Element, t-1)
	tmpD := make([]*pbc.Element, t-1)
	for i, o := range others {
		tmpN[i] = o
		tmpD[i] = pairing.NewZr().Sub(o, cur)
	}
	accumN := PIz(tmpN)
	accumD := PIz(tmpD)
	ret := pairing.NewZr().Div(accumN, accumD)
	return ret
}

func PIz(vals []*pbc.Element) *pbc.Element {
	accum := pairing.NewZr().Set1()
	for _, v := range vals {
		if v != nil {
			accum.MulZn(accum, v)
		}
	}
	return accum
}

// PIgt returns the product of GT inputs
func PIgt(vals []*pbc.Element) *pbc.Element {
	accum := pairing.NewGT().Set1()
	for _, v := range vals {
		if v != nil {
			accum.Mul(accum, v)
		}
	}
	return accum
}

func PIg1(vals []*pbc.Element) *pbc.Element {
	accum := pairing.NewG1().Set1()
	for _, v := range vals {
		if v != nil {
			accum.Mul(accum, v)
		}
	}
	return accum
}

func useless(stub shim.ChaincodeStubInterface) peer.Response {
	nodeID := getNodeID(stub)
	key := "useless" + nodeID
	times, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Can't get useless state: " + err.Error())
	}
	if times == nil {
		times = []byte{1}
	} else {
		times[0] += 1
	}
	err = stub.PutState(key, times)
	if err != nil {
		return shim.Error("Can't put useless state: " + err.Error())
	}
	return shim.Success(times)
}

func main() {
	err := shim.Start(new(FirstTemp))
	if err != nil {
		fmt.Println("Can't start chaincode: " + err.Error())
	}
}
