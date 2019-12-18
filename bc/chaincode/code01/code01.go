/*
 * @author DooQY
 * @createDate 2019/10/25 - 上午11:39
 */

package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type CC struct {
}

func (C CC) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (C CC) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("Hello World!"))
}

func main() {
	err := shim.Start(new(CC))
	if err != nil {
		fmt.Println("Can't start chanicode: " + err.Error())
	}
}
