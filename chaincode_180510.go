/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lpp

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode2 example simple Chaincode implementation
type SimpleChaincode2 struct {
}

func (t *SimpleChaincode2) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Init")
	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// 初始化系统钱包6亿token
	// system_wallet := "02dd5b60206ba8752611c1f564828ddb5111bce5c3ed34cb3e2048aff76fcdfccc"
	// exist, err := checkWalletExisted(stub, system_wallet)
	// if !exist {
	// 	err = stub.PutState(system_wallet, []byte(fmt.Sprint(6e8)))
	// 	if err != nil {
	// 		return shim.Error(err.Error())
	// 	}
	// }

	return shim.Success(nil)
}

func (t *SimpleChaincode2) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Invoke zht")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Make payment of X units from A to B
		return t.invoke(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.query(stub, args)
	} else if function == "add" {
		return t.add(stub, args)
		// 钱包
	} else if function == "walletCreate" {
		return t.walletCreate(stub, args)
		// } else if function == "walletAward" {
		// return t.walletAward(stub, args)
	} else if function == "walletPay" {
		return t.walletPay(stub, args)
	}else if function =="lppInvoke"{
		//lpp
		return t.lppInvoke(stub, args)
	}else if function=="lppCreateAccout"{
		return t.lppCreateccount(stub,args)
	}
	return shim.Error("Invalid invoke function name. Expecting 'invoke' 'delete' 'query' 'walletCreate' 'walletPay' 'lppInvoke'") // 'walletAward'
}
//=======================================lpp==============================================================================
//注册
type LppUser struct {
	UserId       string //用户ID
	PhoneNumber  string //手机号
	UserType     string //用户类型 1.普通用戶,2.企業
	CreateTime   string //创建时间
	CreateUser   string //创建人
	UpdateTime   string //修改时间
	Token        int    //token
}

//交易数据
//眾籌
type CrowdFunding struct {
	ObjectType            string `json:"docType"`         // 类型标示  transaction
	ID                    string `json:"ID"`              // ID
	Name                  string `json:"name"`            // 捐赠的众筹名称
	SpeciesId             string `json:"speciesId"`       // 物种Id
	CrowdfundingId        string `json:"crowdfundingId"`  // 众筹的Id
	Amount                string `json:"amount"`          // 捐赠金额
	CreateTime            string `json:"createTime"`      // 创建时间
}
//冠名
type SpeciesNaming struct {
	ObjectType            string `json:"docType"`         // 类型标示  t_species_naming
	ID                    string `json:"ID"`              // ID
	SpeciesId             string `json:"name"`            // 物种ID
	NamingUserId          string `json:"speciesId"`       // 冠名的userId
	CreateTime            string `json:"crowdfundingId"`  // 创建时间

}
//爱心钻的变化
type LoveDrillChange struct {
	ObjectType            string `json:"docType"`              // 类型标示  LoveDrillChange
	ID                    string `json:"ID"`                   // ID
	Name                  string `json:"name"`                 // 变更名称
	UserId                string `json:"userId"`            // 用户id
	ChangeType            string `json:"changeType"`       // 变更的类型 1、用户签到奖励  2 用户推荐新用户  3 消费购物 4 捐赠样本  5 用户注册 6 捐赠资金
	ChangeId              string `json:"changeId"`                 // '变更的ID，根据type确定， 如果为1 则指向用户签到记录表  如果为2，则指向推荐历史记录表  3 消费购物 则指向订单表  4 指向捐赠奖励指向捐赠记录的Id
	ChangeAmount          string `json:"changeAmount"`            // 变更的爱心钻数量
	OperateUserId         string `json:"operateUserId"`       // 操作人员ID  如果changeType 为 1 ，2，4 的时候为0，表示系统的奖励  3表示用户消费，指向用户的Id
	OperateNote           string `json:"operateNote"`                 // 操作备注
	CreateTime            string `json:"createTime"`            // 创建时间
}
// Transaction makes payment of X units from A to B
func (t *SimpleChaincode2) lppInvoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var _key,_txType,_value string
	_key =args[0]
	_txType =args[1]
    _value =args[2]
    if _txType=="1"{
		data := new(CrowdFunding)
		err := json.Unmarshal([]byte(_value), data)
		if err != nil {

			return shim.Error(err.Error())
		}
		lppUserJSONasBytes, err := json.Marshal(data)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(_key, lppUserJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
    }
	if _txType=="2"{
		data := new(SpeciesNaming)
		err := json.Unmarshal([]byte(_value), data)
		if err != nil {
			return shim.Error(err.Error())
		}
		speciesNamingJSONasBytes, err := json.Marshal(data)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(_key, speciesNamingJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	if _txType=="3"{
		data := new(LoveDrillChange)
		err := json.Unmarshal([]byte(_value), data)
		if err != nil {
			return shim.Error(err.Error())
		}
		loveDrillChangeJSONasBytes, err := json.Marshal(data)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(_key, loveDrillChangeJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
		//将用户token
		userInfoBytes,err :=stub.GetState(data.UserId)
		if err !=nil{
			return shim.Error(err.Error())

		}else if userInfoBytes ==nil{
			return shim.Error("userInfoBytes does not exist")
		}
		lppUser :=LppUser{}
		_amount,err :=strconv.Atoi(data.ChangeAmount)
		if err !=nil{
			return shim.Error(" ")
		}
		if data.ChangeType=="25"{

          lppUser.Token=lppUser.Token -_amount
		}else {
			lppUser.Token=lppUser.Token + _amount
		}
		lppUserJSONasBytes, err := json.Marshal(lppUser)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(lppUser.UserId, lppUserJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}

	return shim.Success(nil)
}


/** Descrption: lpp創建用戶
 *  CreateTime: 2019/09/06 14:24:30
 *  Author: songtieqiang@genomics.cn
 */
func (t *SimpleChaincode2) lppCreateccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key,jsonResp string
	if len(args) !=2{
		return shim.Error("Incorrect number of arguments.")
	}
	key =args[0]
	userExist, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	} else if userExist != nil {
		jsonResp = "{\"Error\":\"Key already  exist: " + key + "\"}"
		return shim.Error(jsonResp)
	}
   //如果用戶不存在直接新增
	lppUser := new(LppUser)
	//时间戳测试
	txTimestamp, tErr := stub.GetTxTimestamp()
	if tErr != nil {
		fmt.Println("Error: get timestamp failed")
	}
	updateTime := time.Unix(txTimestamp.Seconds, 0).Add(time.Hour*time.Duration(8)).Format("2006-01-02 15:04:05")
	fmt.Println("系统的时间戳：" + updateTime)
	lppUser.CreateTime=updateTime
	lppUser.Token=0
	lppUser.UserType="1"
	lppUser.UpdateTime=updateTime
	lppUser.UserId=args[1]

	lppUserJSONasBytes, err := json.Marshal(lppUser)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(key, lppUserJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}
//---------------------------------------------lpp====================================

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode2) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode2) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode2) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)

	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

//added by pgm ,2018 04 02
// Transaction makes payment of X units from A to B
func (t *SimpleChaincode2) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, Aval string // Entities
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	A = args[0]
	Aval = args[1]

	// Write the state back to the ledger
	err = stub.PutState(A, []byte((Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

/** Descrption: 钱包用户创建
 *  CreateTime: 2018/12/11 17:24:53
 *      Author: zhoutong@genomics.cn
 */
func (t *SimpleChaincode2) walletCreate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	exist, err := checkWalletExisted(stub, A)
	if err != nil {
		return shim.Error(err.Error())
	}

	if exist {
		return shim.Error(fmt.Sprintf("Wallet %v has existed", A))
	}

	err = stub.PutState(A, []byte(fmt.Sprint(0.0)))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

/** Descrption: 钱包充钱，天赐之
 *  CreateTime: 2018/12/11 17:24:53
 *      Author: zhoutong@genomics.cn
 */
// func (t *SimpleChaincode2) walletAward(stub shim.ChaincodeStubInterface, args []string) pb.Response {
// 	if len(args) != 2 {
// 		return shim.Error("Incorrect number of arguments. Expecting 2")
// 	}

// 	A := args[0]

// 	Aval, err := getWalletBalance(stub, A)
// 	if err != nil {
// 		return shim.Error(fmt.Sprintf("{\"Error\":\"parse wallet %v amount error: %v\"}", A, err))
// 	}

// 	X, err := strconv.ParseFloat(args[1], 64)
// 	if err != nil {
// 		return shim.Error("Invalid wallet amount, expecting a float value")
// 	}

// 	Aval += X
// 	err = stub.PutState(A, []byte(fmt.Sprint(Aval)))
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
// 	return shim.Success(nil)
// }

/** Descrption: 获取钱包余额
 *  CreateTime: 2018/12/12 11:29:06
 *      Author: zhoutong@genomics.cn
 */
func getWalletBalance(stub shim.ChaincodeStubInterface, addr string) (Aval float64, err error) {
	Avalbytes, err := stub.GetState(addr)
	if err != nil {
		err = fmt.Errorf("Failed to get state for %v: %v", addr, err)
		return
	}

	if Avalbytes == nil {
		err = fmt.Errorf("Nil amount for %v", addr)
		return
	}

	Aval, err = strconv.ParseFloat(string(Avalbytes), 64)
	if err != nil {
		err = fmt.Errorf("Parse wallet %v amount error: %v", addr, err)
		return
	}
	return
}

/** Descrption: 确认钱包存在
 *  CreateTime: 2018/12/12 11:28:30
 *      Author: zhoutong@genomics.cn
 */
func checkWalletExisted(stub shim.ChaincodeStubInterface, addr string) (exist bool, err error) {
	Avalbytes, err := stub.GetState(addr)
	if err != nil {
		err = fmt.Errorf("Failed to get state for %v: %v", addr, err)
		return
	}

	if Avalbytes != nil {
		exist = true
		return
	}
	return
}

/** Descrption: 钱包付款，转同样的金额给多人
 *  CreateTime: 2018/12/11 17:26:06
 *      Author: zhoutong@genomics.cn
 */
func (t *SimpleChaincode2) walletPay(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// senderAddr,amount,recieverAddrs
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting at least 3")
	}

	A := args[0]
	Aval, err := getWalletBalance(stub, A)
	if err != nil {
		return shim.Error(err.Error())
	}

	X, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("Invalid wallet amount, expecting a float value")
	}

	// 防止转账变收账
	if X <= 0 {
		return shim.Error("Invalid wallet amount, expecting a positive value")
	}

	recievers := args[2:]
	cost := X * float64(len(recievers))

	// 余额不足
	if Aval < cost {
		return shim.Error("Balance is not enough")
	}

	Aval -= cost
	err = stub.PutState(A, []byte(fmt.Sprint(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// 此时新Aval还没写入数据库，GetState(A)仍然是原值。需等待合约成功才会正式写入值

	var recvCounts = make(map[string]int) // 收账人出现次数
	for _, B := range recievers {
		// 不允许转账给自己
		if B == A {
			return shim.Error("Paying to yourself is not allowed")
		}

		recvCounts[B]++
	}

	for B, cnt := range recvCounts {
		Bval, err := getWalletBalance(stub, B)
		if err != nil {
			return shim.Error(err.Error())
		}

		// B出现多次时，表示付给B多倍金额
		Bval += X * float64(cnt)
		err = stub.PutState(B, []byte(fmt.Sprint(Bval)))
		if err != nil {
			shim.Error(err.Error())
		}
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SimpleChaincode2))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
