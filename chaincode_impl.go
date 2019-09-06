package main

import (
	"bytes"
	//"chaincode/go/jifen/src/util"
	//"github.com/hyperledger/fabric/examples/chaincode/go/jifen/src/util"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type PointsChaincode struct {
}
type SuperObject2 interface{}

//积分交易记录对象
type PointsTransaction struct {
	TransId        string //积分交易ID
	RolloutAccount string //转出账户
	RollinAccount  string //转入账户
	TransAmount    string //交易积分数量
	Description    string //描述
	TransferTime   string //交易时间
	TransferType   string //交易类型
	CreateTime     string //创建时间
	CreateUser     string //创建人
	UpdateTime     string //修改时间
	UpdateUser     string //修改人
	OperFlag       string // 操作标积 0-新增，1-修改，2-删除
}

//积分交易明细对象
type PointsTransactionDetail struct {
	DetailId         string //逐笔明细流水号
	SourceDetailId   string //来源流水号
	TransId          string //积分交易ID
	RolloutAccount   string //转出账户
	RollinAccount    string //转入账户
	TransAmount      string //交易积分数量
	ExpireTime       string //过期时间
	ExtRef           string //外部引用
	Status           string // 状态  0-冻结，1-正常
	CurBalance       string //当笔积分剩余数量
	TransferTime     string //交易时间
	CreditCreateTime string //授信创建时间
	CreditParty      string //授信方账户
	Merchant         string // 商户账户
	CreateTime       string //创建时间
	CreateUser       string //创建人
	UpdateTime       string //修改时间
	UpdateUser       string //修改人
	OperFlag         string // 操作标积 0-新增，1-修改，2-删除
}

//授信的传参
type CreditPointsTransData struct {
	PointsTransaction       *PointsTransaction
	PointsTransactionDetail *PointsTransactionDetail
}

//消费的传参
type ConsumePointsTransData struct {
	PointsTransaction           *PointsTransaction
	PointsTransactionDetailList []*PointsTransactionDetail
}

//承兑参数传递
type AccpetPointsTransData struct {
	PointsTransaction           []*PointsTransaction
	PointsTransactionDetailList []*PointsTransactionDetail
}

//初始化创建表
func (t *PointsChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	log.Println("Init method.........")
	return shim.Success(nil)
}

//调用方法
func (t *PointsChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//defer util.End(util.Begin("Invoke"))
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)
	if function == "CreditPoints" { // 授信积分
		//log.Println("CreditPoints,args = " + args[0])
		return t.CreditPoints(stub, args)
	} else if function == "queryPointsTransaction" { //cesi
		return t.queryPointsTransaction(stub, args)
	} else if function == "readPointsTransaction" {
		return t.readPointsTransaction(stub, args)
	} else if function == "ConsumePoints" { //消费
		return t.ConsumePoints(stub, args)
	} else if function == "InitData" { //注册用户

		return t.InitData(stub, args)
	} else if function == "AccpetPoints" { // 承兑积分

		return t.AccpetPoints(stub, args)
	}

	return shim.Error("Received unknown function invocation")
}

//查询单笔任何交易的功能前提用它的key
// ===============================================
// readPointsTransaction - read a marble from chaincode state
// ===============================================
func (t *PointsChaincode) readPointsTransaction(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var TransId, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	TransId = args[0]
	valAsbytes, err := stub.GetState(TransId) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + TransId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"TransId does not exist: " + TransId + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

//测试查询交易表的记录
func (t *PointsChaincode) queryPointsTransaction(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]
	fmt.Println("query: " + queryString) //query
	//queryResults, err := util.GetQueryResultForQueryString(stub, queryString)
	queryResults, err := GetQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

//查询功能
func GetQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func ParseJsonAndDecode(data SuperObject2, args []string) error {
	fmt.Println(data)
	//base64解码
	arg, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		log.Println("ParseJson base64 decode error.")
		return err
	}
	log.Println("data after decode log:" + string(arg[:]))
	fmt.Println("data after decode fmt:" + string(arg[:]))

	//解析数据
	err = json.Unmarshal(arg, data)
	if err != nil {
		log.Println("ParseJson json Unmarshal error.")
		return err
	}
	fmt.Println("----------------------------------")
	fmt.Println(data)
	fmt.Println("Parse json is ok.")

	return nil
}

//授信积分
func (t *PointsChaincode) CreditPoints(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	// 解析传入数据
	creditObject := new(CreditPointsTransData)
	//用base64传进来的参数解码并转化为对象
	err := ParseJsonAndDecode(creditObject, args)
	if err != nil {
		log.Println("Error occurred when parsing json")
		return shim.Error(err.Error())
	}
	//直接传的参数
	//	err = json.Unmarshal(args, creditObject)
	//	if err != nil {
	//		log.Println("ParseJson json Unmarshal error.")
	//		return shim.Error(err.Error())
	//	}

	//账户信息表更新？需不需要账户表？

	// 积分交易表增加记录
	if creditObject.PointsTransaction.OperFlag == "0" {
		//插入积分交易表
		fmt.Println("creditObject.PointsTransaction value=======start")
		fmt.Println(creditObject.PointsTransaction)
		fmt.Println("creditObject.PointsTransaction value=======end")

		pointsTransactionJSONasBytes, err := json.Marshal(creditObject.PointsTransaction)
		if err != nil {
			return shim.Error(err.Error())
		}

		// === Save PointsTransaction to state ===

		err = stub.PutState(creditObject.PointsTransaction.TransId, pointsTransactionJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}

	} else {
		log.Println("Flag is not 0,that's why we do nothing for table points_transation")
	}
	// 积分交易明细表增加记录
	if creditObject.PointsTransactionDetail.OperFlag == "0" {

		fmt.Println("creditObject==== ointsTransactionDetail value=======start")
		fmt.Println(creditObject.PointsTransactionDetail)
		fmt.Println("creditObject==== ointsTransactionDetail value=======end")
		PointsTransactionDetailJSONasBytes, err := json.Marshal(creditObject.PointsTransactionDetail)
		if err != nil {
			return shim.Error(err.Error())
		}
		// === Save PointsTransactionDetail to state ===
		err = stub.PutState(creditObject.PointsTransactionDetail.DetailId, PointsTransactionDetailJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}

	} else {
		log.Println("Flag is not 0,that's why we do nothing for table points_transation_detail")
	}
	log.Println("CreditPoints success.")

	return shim.Success(nil)
}

//消费积分:
func (t *PointsChaincode) ConsumePoints(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// parse data
	data := new(ConsumePointsTransData)
	err := ParseJsonAndDecode(data, args)
	if err != nil {
		log.Println("Error occurred when parsing json")
		return shim.Error(err.Error())
	}

	transAmount := data.PointsTransaction.TransAmount // transfer amount

	amount, err := strconv.ParseInt(transAmount, 10, 64)
	if err != nil {
		log.Println("Error occurred when converting amount string to int")
		return shim.Error(err.Error())
	}
	if amount == 0 {
		errorMsg := "transfer amount is zero,so no need to do any transaction."
		return shim.Error(errorMsg)
	}

	//=============================================这里主要是为了校验===================================
	// update points transaction table.
	pointsObj := data.PointsTransaction
	trans, _ := strconv.ParseInt(pointsObj.TransAmount, 10, 64) // transaction points amount

	// define a variable to store exchange amounts from all update amount.
	var totalUpdate int64

	// upate points transaction detail table
	for i := 0; i < len(data.PointsTransactionDetailList); i++ {
		detail := data.PointsTransactionDetailList[i]

		if detail.OperFlag == "0" { //新增

			// current balance of last transaction detail

			curBalance1, _ := strconv.ParseInt(QueryPointsDetailCurBalanceByKey(stub, detail.SourceDetailId), 10, 64)

			// check if last transaction detail exists
			result := CheckPointsDetailExist(stub, detail.SourceDetailId)
			if !result {
				var errorMsg = "Table Points_Transation_Detail: specified record doesn't exist,detail.SourceDetailId = " + detail.SourceDetailId
				return shim.Error(errorMsg)
			}

			// check if remaining points of last transaction detail is enough.
			transPoints, _ := strconv.ParseInt(detail.TransAmount, 10, 64)
			if transPoints > curBalance1 {
				log.Println("transPoints ->" + strconv.FormatInt(transPoints, 10))
				log.Println("curBalance1 ->" + strconv.FormatInt(curBalance1, 10))
				var errorMsg = "Current balance of last transaction detail is not enough to pay,detail.SourceDetailId = " + detail.SourceDetailId
				return shim.Error(errorMsg)
			}
		} else {

			// current balance of last transaction detail
			curBalance2, _ := strconv.ParseInt(QueryPointsDetailCurBalanceByKey(stub, detail.DetailId), 10, 64)
			log.Println("curBalance2=" + strconv.FormatInt(curBalance2, 10))
			temp, _ := strconv.ParseInt(detail.CurBalance, 10, 64)
			log.Println("temp=" + strconv.FormatInt(temp, 10))

			// compute exchange amount
			changeAmount := curBalance2 - temp
			log.Println("changeAmount=" + strconv.FormatInt(changeAmount, 10))
			totalUpdate += changeAmount
			log.Println("totalUpdate=" + strconv.FormatInt(totalUpdate, 10))
		}
	}

	log.Println("trans=" + strconv.FormatInt(trans, 10))

	if trans != totalUpdate {
		errMsg := "Invalid data, pls. check if this request has been tampered"
		return shim.Error(errMsg)
	}

	// performing insert or update
	log.Println("Performing insert/update start............")

	InsertPointsTransation(stub, pointsObj)
	for i := 0; i < len(data.PointsTransactionDetailList); i++ {
		detailObject := data.PointsTransactionDetailList[i]
		if detailObject.OperFlag == "0" {
			InsertPointsTransationDetail(stub, detailObject)
		} else {
			UpdatePointsTransationDetail(stub, detailObject)
		}
	}
	log.Println("Performing insert/update end............")

	log.Println("ConsumePoints success.")
	return shim.Success(nil)
}

//承兑积分：这里不做验证了 校验的直接把消费积分的校验拿来用即可
func (t *PointsChaincode) AccpetPoints(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	// 解析传入数据
	data := new(AccpetPointsTransData)
	err := ParseJsonAndDecode(data, args)
	if err != nil {
		log.Println("Error occurred when parsing json")
		return shim.Error(err.Error())
	}
	for i := 0; i < len(data.PointsTransaction); i++ {
		//如果标识符为0就对账户表新增否则修改
		if data.PointsTransaction[i].OperFlag == "0" {
			InsertPointsTransation(stub, data.PointsTransaction[i])
		} else {
			return shim.Error("cuowu")
		}
	}
	for i := 0; i < len(data.PointsTransactionDetailList); i++ {
		if data.PointsTransactionDetailList[i].OperFlag == "0" {
			InsertPointsTransationDetail(stub, data.PointsTransactionDetailList[i])
		} else {
			UpdatePointsTransationDetail(stub, data.PointsTransactionDetailList[i])
		}
	}
	log.Println("AccpetPoints success.")
	return shim.Success(nil)
}

func main() {

	err := shim.Start(new(PointsChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}

}

//功能区
//根据主键查询账户余额
func QueryPointsDetailCurBalanceByKey(stub shim.ChaincodeStubInterface, detailId string) string {

	PointsDetailCurBalanceAsBytes, err := stub.GetState(detailId)
	if err != nil {
		return "0"
	} else if PointsDetailCurBalanceAsBytes == nil {
		return "0"
	}

	PointsTransactionDetail := PointsTransactionDetail{}
	err = json.Unmarshal(PointsDetailCurBalanceAsBytes, &PointsTransactionDetail) //unmarshal it aka JSON.parse()
	if err != nil {
		return "0"
	}

	return PointsTransactionDetail.CurBalance

}

//校验是否存在这笔积分
func CheckPointsDetailExist(stub shim.ChaincodeStubInterface, detailId string) bool {

	PointsDetailCurBalanceAsBytes, err := stub.GetState(detailId)
	if err != nil {
		return false
	} else if PointsDetailCurBalanceAsBytes == nil {
		return false
	}
	return true

}

//插入功能交易明细
func InsertPointsTransationDetail(stub shim.ChaincodeStubInterface, pointsDetail *PointsTransactionDetail) pb.Response {

	fmt.Println("=======insert=======")
	PointsTransactionDetailJSONasBytes, err := json.Marshal(pointsDetail)
	if err != nil {
		return shim.Error(err.Error())
	}
	// === Save PointsTransactionDetail to state ===
	err = stub.PutState(pointsDetail.DetailId, PointsTransactionDetailJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

//插入交易
func InsertPointsTransation(stub shim.ChaincodeStubInterface, points *PointsTransaction) pb.Response {

	fmt.Println("=======insert=======")
	PointsTransactionDetailJSONasBytes, err := json.Marshal(points)
	if err != nil {
		return shim.Error(err.Error())
	}
	// === Save PointsTransaction to state ===
	err = stub.PutState(points.TransId, PointsTransactionDetailJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

//更新功能
func UpdatePointsTransationDetail(stub shim.ChaincodeStubInterface, pointsDetail *PointsTransactionDetail) pb.Response {

	fmt.Println("=======update=======")

	//传进来的值有问题
	PointsTransactionDetailAsBytes, err := stub.GetState(pointsDetail.DetailId)
	if err != nil {
		return shim.Error("Failed to get PointsTransactionDetailAsBytes:" + err.Error())
	} else if PointsTransactionDetailAsBytes == nil {
		return shim.Error("PointsTransactionDetailAsBytes does not exist")
	}

	PointsTransactionDetailQukuai := PointsTransactionDetail{}
	err = json.Unmarshal(PointsTransactionDetailAsBytes, &PointsTransactionDetailQukuai) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	PointsTransactionDetailQukuai.CurBalance = pointsDetail.CurBalance //change the owner

	PointsTransactionDetailJSONasBytes, err := json.Marshal(PointsTransactionDetailQukuai)
	if err != nil {
		return shim.Error(err.Error())
	}
	// === Save PointsTransactionDetail to state ===
	err = stub.PutState(pointsDetail.DetailId, PointsTransactionDetailJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

//注册用户================================================================

type Account struct {
	AccountId      string //账户ID
	UserId         string // 账户所属用户ID
	AccountBalance string // 账户积分剩余数量
	AccountTypeId  string // 账户类型ID
	OperFlag       string // 操作标积 0-新增，1-修改，2-删除
	CreateTime     string //创建时间
	CreateUser     string //创建人
	UpdateTime     string //修改时间
	UpdateUser     string //修改人
}

type AccountType struct {
	AccountTypeId string // 账户类型ID
	Describe      string // 描述
	CreateTime    string //创建时间
	CreateUser    string //创建人
	UpdateTime    string //修改时间
	UpdateUser    string //修改人
}
type InitTableData struct {
	PointsUser []*PointsUser
	Account    []*Account
}

//注册
type PointsUser struct {
	UserId       string //用户ID
	UserName     string //用户名称
	UserPassword string //用户密码
	PhoneNumber  string //手机号
	UserType     string //用户类型 0.授信，1.商户,2.会员
	CreateTime   string //创建时间
	CreateUser   string //创建人
	UpdateTime   string //修改时间
	UpdateUser   string //修改人
}


//初始化数据
func (t *PointsChaincode) InitData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	data := new(InitTableData)
	err := ParseJsonAndDecode(data, args)
	if err != nil {
		log.Println("Error occurred when parsing json")
		return shim.Error(err.Error())
	}
	for i := 0; i < len(data.PointsUser); i++ {
		InsertPointsUser(stub, data.PointsUser[i])
	}
	for j := 0; j < len(data.Account); j++ {
		InsertAccount(stub, data.Account[j])
	}
	log.Println("InitData success.")
	return shim.Success(nil)
}
func InsertPointsUser(stub shim.ChaincodeStubInterface, PointsUser *PointsUser) pb.Response {

	fmt.Println("=======insert=======")
	PointsUserDetailJSONasBytes, err := json.Marshal(PointsUser)
	if err != nil {
		return shim.Error(err.Error())
	}
	// === Save PointsTransactionDetail to state ===
	err = stub.PutState(PointsUser.UserId, PointsUserDetailJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
func InsertAccount(stub shim.ChaincodeStubInterface, Account *Account) pb.Response {

	fmt.Println("=======insert=======")
	AccountJSONasBytes, err := json.Marshal(Account)
	if err != nil {
		return shim.Error(err.Error())
	}
	// === Save PointsTransactionDetail to state ===
	err = stub.PutState(Account.AccountId, AccountJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
