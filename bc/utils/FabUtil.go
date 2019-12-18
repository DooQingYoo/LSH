/*
 * @author DooQY
 * @createDate 2019/11/9 - 下午1:31
 */

package utils

import (
	"errors"
	"fmt"
	"github.com/Nik-U/pbc"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"strings"
)

var SDK *fabsdk.FabricSDK
var CommClients []*channel.Client
var NodeNumber int

func init() {
	var client *channel.Client
	var err error
	SDK, err = fabsdk.New(config.FromFile("/home/forsim/go/LSH/bc/config.yaml"))
	if err != nil {
		panic("Can't start sdk: " + err.Error())
	}
	CommClients = append(CommClients, nil)
	client, err = SwitchOrg(1)
	if err != nil {
		panic("Can't start client" + strconv.Itoa(1) + " : " + err.Error())
	}
	CommClients = append(CommClients, client)
	NodeNumber, err = GetNodeNumber()
	if err != nil {
		panic("can't get node number: " + err.Error())
	}
	for i := 2; i <= NodeNumber; i++ {
		client, err = SwitchOrg(i)
		if err != nil {
			panic("Can't start client" + strconv.Itoa(1) + " : " + err.Error())
		}
		CommClients = append(CommClients, client)
	}
}

func GetG() (*pbc.Element, error) {
	gbytes, err := queryPub("g")
	if err != nil {
		return nil, err
	}
	if gbytes == nil {
		return nil, errors.New("No g is initialized ")
	}
	g := pairing.NewG1().SetBytes(gbytes)
	return g, nil
}

func GetNodeNumber() (int, error) {
	nnbytes, err := queryPub("NodeNumber")
	if err != nil {
		return 0, err
	}
	if nnbytes == nil {
		return 0, errors.New("NodeNumber is not saved in the ledger")
	}
	n := int(nnbytes[0])
	if n < 2 || n > 20 {
		return 0, errors.New("Ledger has saved the wrong number of node: " + strconv.Itoa(n))
	}
	return n, nil
}

func GetG1() (*pbc.Element, error) {
	g1bytes, err := queryPub("g1")
	if err != nil {
		return nil, err
	}
	if g1bytes == nil {
		return nil, errors.New("No g1 is initialized ")
	}
	g1 := pairing.NewG1().SetBytes(g1bytes)
	return g1, nil
}

func queryPub(key string) ([]byte, error) {
	queryProposal := channel.Request{
		ChaincodeID: "mycc",
		Fcn:         "queryBytes",
		Args:        [][]byte{[]byte(key)},
	}
	response, err := CommClients[1].Execute(queryProposal, channel.WithTargetEndpoints("peer1.LSH1"))
	if err != nil {
		return nil, errors.New("Query blockchain error: " + err.Error())
	}
	if response.TxValidationCode == peer.TxValidationCode_VALID {

		return response.Payload, nil
	}
	return nil, errors.New("Query blockchain fail, Transaction invalid: " + strconv.Itoa(int(response.TxValidationCode)))
}

func Dispatch(org int, ch chan bool) {
	requestProposal := channel.Request{
		ChaincodeID: "mycc",
		Fcn:         "dispatch",
	}
	response, err := CommClients[org].Execute(requestProposal, channel.WithTargetEndpoints("peer1.LSH"+strconv.Itoa(org)))
	if err != nil {
		fmt.Println(err)
		ch <- false
		return
	}
	fmt.Println(string(response.Payload))
	ch <- true
}

func Broadcast(org int, ch chan bool) {
	requestProposal := channel.Request{
		ChaincodeID: "mycc",
		Fcn:         "broadcast",
	}
	response, err := CommClients[org].Execute(requestProposal, channel.WithTargetEndpoints("peer1.LSH"+strconv.Itoa(org)))
	if err != nil {
		fmt.Println(err)
		ch <- false
		return
	}
	pki := pairing.NewGT().SetBytes(response.Payload)
	fmt.Println("pki" + strconv.Itoa(org) + "=" + pki.String())
	ch <- true
}

func Assemble(ch chan bool) {
	requestProposal := channel.Request{
		ChaincodeID: "mycc",
		Fcn:         "assemble",
	}
	response, err := CommClients[1].Execute(requestProposal, channel.WithTargetEndpoints("peer1.LSH1"))
	if err != nil {
		fmt.Println(err)
		ch <- false
		return
	}
	publicE := pairing.NewGT().SetBytes(response.Payload)
	fmt.Println("publicE = " + publicE.String())
	ch <- true
}

func GetKLH(org int, UserID, Ak, Bk, Ck *pbc.Element, HashAttr []*pbc.Element, ch chan bool) {
	// 根据使用哪个节点生成相应的channel Client
	client := CommClients[org]

	// 发送请求
	var args [][]byte
	args = append(args, UserID.Bytes())
	args = append(args, Ak.Bytes())
	args = append(args, Bk.Bytes())
	args = append(args, Ck.Bytes())
	for _, v := range HashAttr {
		args = append(args, v.Bytes())
	}
	request := channel.Request{
		ChaincodeID: "mycc",
		Fcn:         "getKAndL",
		Args:        args,
	}
	response, err := client.Execute(request, channel.WithTargetEndpoints("peer1.LSH"+strconv.Itoa(org)))
	if err != nil {
		fmt.Println(err)
		ch <- false
		return
	}
	fmt.Println("节点"+strconv.Itoa(org)+"的结果为:\n"+string(response.Payload))
	ch <- true
}

func GetFinalKey(UserID *pbc.Element, selectedNodes []byte, ch chan bool) {

	request := channel.Request{
		ChaincodeID: "mycc",
		Fcn:         "getFinalKey",
		Args:        [][]byte{UserID.Bytes(), selectedNodes},
	}
	response, err := CommClients[1].Execute(request, channel.WithTargetEndpoints("peer1.LSH1"))
	if err != nil {
		fmt.Println(err)
		ch <- false
	}
	key := string(response.Payload)
	fmt.Println("最终结果: ")
	keys := strings.Split(key, ";")
	fmt.Println("K = " + keys[0])
	fmt.Println("L = " + keys[1])
	fmt.Println("{ Kx } = ")
	for i := 2; i < len(keys); i++ {
		fmt.Println(keys[i])
	}
	ch <- true
}

func UselessFunc(org int, ch chan bool) {
	client := CommClients[org]
	request := channel.Request{
		ChaincodeID: "mycc",
		Fcn:         "useless",
	}
	_, err := client.Execute(request, channel.WithTargetEndpoints("peer1.LSH"+strconv.Itoa(org)))
	if err != nil {
		fmt.Println(err)
		ch <- false
		return
	} else {
		ch <- true
	}
}

func SwitchOrg(org int) (*channel.Client, error) {
	orgID := "LSH" + strconv.Itoa(org)
	context := SDK.ChannelContext("mychannel", fabsdk.WithOrg(orgID), fabsdk.WithUser("User1"))
	if context == nil {
		panic("can't get channel context!")
	}
	client, err := channel.New(context)
	if err != nil {
		return &channel.Client{}, err
	}
	return client, nil
}
