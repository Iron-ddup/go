package main

//引入类库
import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	NOSTROTYPE = "nostro"
	TXTYPE     = "txinstr"
)

//NOSTROjson结构体
type TdNoStro struct {
	ObjectType string `json:"docType"`   // 类型标示  nostro
	ACTID      string `json:"actid"`     // 往来账户ID
	BKCODE     string `json:"bkcode"`    // 银行号
	CLRBKCDE   string `json:"clrbkcde"`  // 资金清算银行号
	CURCDE     string `json:"curcde"`    // 货币码
	NOSTROBAL  string `json:"nostrobal"` // Nostro余额
}

//TxInstrjson结构体
type TxInstr struct {
	ObjectType  string `json:"docType"`     // 类型标示  txinstr
	INSTRID     string `json:"instrid"`     // 交易指示ID
	BKCODE      string `json:"bkcode"`      // 转出银行号
	ACTNOFROM   string `json:"actnofrom"`   // 转出账号
	CLRBKCDE    string `json:"clrbkcde"`    // 转入银行号
	ACTNOTO     string `json:"actnoto"`     // 转入账号
	TXAMT       string `json:"txamt"`       // 交易金额
	CURCDE      string `json:"curcde"`      // 货币码
	TXNDAT      string `json:"txndat"`      // 交易日期
	TXNTIME     string `json:"txntime"`     // 交易时间
	COMPST      string `json:"compst"`      // 交易完成状态 Completion Status(N-new added,P – pending, S – success，R-rejected)
	RTCDE       string `json:"rtcde"`       // 交易状态代码 Completion Status Code (returned from core banking system)
	SECRETKEYID string `json:"secretkeyid"` //公钥加密ID
}

//公钥上链结构体
type PublicKeyStruct struct {
	ObjectType           string `json:"docType"`              // 类型标示  PublicKey
	Pubid                string `json:"Pubid"`                // id
	Role                 string `json:"Role"`                 // 角色
	Dt                   string `json:"Date"`                 // 日期
	OrganizationCertName string `json:"OrganizationCertName"` // 组织证书名字
	Status               string `json:"Status"`               // 状态
	PublicKey            string `json:"PublicKey"`
	OrganizationName     string `json:"OrganizationName"` // 组织名字
}

//vote
type Vote struct {
	ObjectType           string `json:"docType"`              // 类型标示  Vote
	Voteid               string `json:"Voteid"`               // id
	OrganizationName     string `json:"OrganizationName"`     // 申请监管机构名
	VoteOrganization     string `json:"VoteOrganization"`     //投票机构
	DeadlineDate         string `json:"DeadlineDate"`         //截至日期
	VoteDate             string `json:"VoteDate"`             //投票日期
	BanktoMonitor        string `json:"BanktoMonitor"`        // 需要被监管的银行
	InitiateOrganization string `json:"InitiateOrganization"` //发起机构
	Opinion              string `json:"Opinion"`              // 意见Y/N
	Ruleid               string `json:"Ruleid"`               // 规则id
}

//rule
type Rule struct {
	ObjectType      string `json:"docType"`         // 类型标示  Rule
	Ruleid          string `json:"Ruleid"`          // id
	DeadlineHourNum string `json:"DeadlineHourNum"` // 延长多久过期
	AgreedNum       string `json:"AgreedNum"`       // 达到多少机构同意

}

type SimpleChaincode struct {
}

//初始化chaincode
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

//路由
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)
	fmt.Println(args)

	//获取机构的证书名字
	Organization := t.GetCertificate(stub, args)
	fmt.Println("--------Certificate-------------------")
	fmt.Println(Organization)

	//查询处理=====如果是查询所有人都可以查：数据都是加密处理的
	if function == "QueryTdNoStro" {
		return t.QueryTdNoStro(stub, args)
	} else if function == "QueryTransaction" {
		return t.QueryTransaction(stub, args)
	} else if function == "QueryTxInstrByActnoto" {
		return t.QueryTxInstrByActnoto(stub, args)
	} else if function == "QueryTxInstrByActnofrom" {
		return t.QueryTxInstrByActnofrom(stub, args)
	} else if function == "QueryTxInstrByClrbkcde" {
		return t.QueryTxInstrByClrbkcde(stub, args)
	} else if function == "QueryTxInstrByBkcode" {
		return t.QueryTxInstrByBkcode(stub, args)
	} else if function == "QueryByString" { //test query
		return t.QueryByString(stub, args)
	} else if function == "BankQuery" {
		return t.BankQuery(stub, args)
	} else if function == "UserQuery" {
		return t.UserQuery(stub, args)
	} else if function == "QueryPublicKey" { //查询公钥
		return t.QueryPublicKey(stub, args)
	} else if function == "QueryVotedata" { //查询投票具体数据/所有 和已投的
		return t.QueryVotedata(stub, args)
	} else if function == "QuerySecretKey" { //查询加密密钥
		return t.QuerySecretKey(stub, args)
	} else if function == "InitPublicKey" { //初始化公钥只能一条一条的插入
		return t.InitPublicKey(stub, args, Organization)
	} else if function == "QuerySinglebyId" { //可以查询任何单条记录
		return t.QuerySinglebyId(stub, args)
	} else if function == "InitNoStro" { //初始化nostro
		return t.InitNoStro(stub, args)
	} else if function == "BankAndMonitor" { //bank和监管机构
		return t.BankAndMonitor(stub, args)
	} else if function == "InitPublicKey2" { //可以同时上传多条公钥
		return t.InitPublicKey2(stub, args)
	} else if function == "AddOrganizationPublicKey2Tx3" { //增加监管机构后对监管的银行交易监管
		return t.AddOrganizationPublicKey2Tx3(stub, args)
	} else if function == "UpdateSecretKey2" { //修改公共密钥的状态：泄密后不用
		return t.UpdateSecretKey2(stub, args)
	} else if function == "QuerySecretKey2" { //查询公钥通过字符串的方式banks:BankABankB/BankBBankA(不建议使用)
		return t.QuerySecretKey2(stub, args)
	} else if function == "QueryMonitor2Bank" { //查询监管机构监管几家银行
		return t.QueryMonitor2Bank(stub, args)
	} else if function == "QueryBanks" { //查询链上的银行机构（用于投票）
		return t.QueryBanks(stub, args)
	} else if function == "QuerySecretKeyAll" { //查询所有的密钥
		return t.QuerySecretKeyAll(stub, args)
	}

	role, orgName := GetRoleAndOrgName(stub, Organization) //通过机构证书的名字，查询到他的"角色"
	fmt.Println("role:" + role)
	fmt.Println("orgName:" + orgName)
	if role == "" || orgName == "" {
		return shim.Error("公钥没有上链")
	}
	//假设有银行两家，监管机构1家，管理员1家：CNCB CNCBI  Police Manager
	//链上查询：处理办法先用获得的“证书名”得到它的“角色” ，角色role（机构（Organization），银行（Bank）,管理员（Manager））
	//所以在操作之前必须所有机构的（公钥，角色，机构证书名。。。）上链
	//现在不需要管理员
	if role == "Monitor" { //机构
		switch function {
		case "UpdateSecretKey2":
			return t.UpdateSecretKey2(stub, args) //修改产生的密钥
		case "IncreaseSecretKey":
			return t.IncreaseSecretKey(stub, args) //新增产生的密钥
		case "Vote":
			return t.Vote(stub, args, orgName) //发起投票
		case "AddOrganizationPublicKey2Tx3":
			return t.AddOrganizationPublicKey2Tx3(stub, args) //新机构上公钥
		case "QueryVoteOpinion":
			return t.QueryVoteOpinion(stub, orgName) //查询自己需要投票的和这个投票是否完成
		}
	} else if role == "Bank" { //银行 查询：保证每家银行只能查到自己的转入转出（需要？）不需要：反正是加密数据
		switch function {
		case "TxInstr":
			return t.TxInstr(stub, args) //交易产生
		case "UpdateCompst":
			return t.UpdateCompst(stub, args) //更改交易状态
		case "UpdateNoStro":
			return t.UpdateNoStro(stub, args) //修改NoStro
		case "IncreaseSecretKey":
			return t.IncreaseSecretKey(stub, args) //新增产生的密钥
		case "UpdateSecretKey2":
			return t.UpdateSecretKey2(stub, args) //修改产生的密钥
		case "Vote":
			return t.Vote(stub, args, orgName) //投票这个触发 目前暂定等讨论
		case "CreateVote":
			return t.CreateVote(stub, args) //创建投票
		case "CreateRule":
			return t.CreateRule(stub, args) //创建规则
		case "QueryVoteOpinion":
			return t.QueryVoteOpinion(stub, orgName) //查询自己需要投票的和这个投票是否完成
		case "QuerySecretKey3":
			return t.QuerySecretKey3(stub, orgName) //查询相应银行用到的密钥
		}
	}
	errMsg := fmt.Sprintln("Received unknown function invocation: [function:", function, ", args:", args, "].")
	return shim.Error(errMsg)

}

//查询自己需要投票的和这个投票是否完成
//传进来的是银行
func (t *SimpleChaincode) QueryVoteOpinion(stub shim.ChaincodeStubInterface, orgName string) pb.Response {
	fmt.Println("Invoke  QueryVoteOpinion")
	var jsonResp string
	fmt.Println(orgName)
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"Vote\",\"VoteOrganization\":\"%s\"}}", orgName)
	fmt.Println(queryString)
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error("This resultsIterator does't already exists1 ")
	}
	defer resultsIterator.Close()
	//定义一个字节输出流
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("This publickey does't already exists2 ")
		}

		//fmt.Println(string(queryResponse.Value))
		fmt.Println(string(queryResponse.Value))
		//将得到的Value解析为结构体用来操作
		VoteStruct := new(Vote)
		err = json.Unmarshal(queryResponse.Value, VoteStruct)
		createId := VoteStruct.Voteid
		BankArray := strings.Split(createId, ":") //得到传进来的参数Split
		voteValAsbytes, err := stub.GetState(BankArray[0])
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to get state for " + BankArray[0] + "\"}"
			return shim.Error(jsonResp)
		} else if voteValAsbytes == nil {
			jsonResp = "{\"Error\":\"Marble does not exist: " + BankArray[0] + "\"}"
			return shim.Error(jsonResp)
		}
		//顺便把公钥也返回====start===
		//申请机构的名字
		shengqinBankName := VoteStruct.OrganizationName
		shengqinBankPub := GetNewOrgPub(stub, shengqinBankName)
		//================end=========
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"CreateVote\":")

		buffer.WriteString(string(voteValAsbytes))

		buffer.WriteString(", \"Vote\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString(", \"PublicKey\":")
		buffer.WriteString(shengqinBankPub)
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true

	}
	buffer.WriteString("]")
	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())
	return shim.Success(buffer.Bytes())
}

//单独通过机构名获取公钥
func GetNewOrgPub(stub shim.ChaincodeStubInterface, organization string) string {
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"PublicKey\",\"Status\":\"latest\",\"OrganizationName\":\"%s\"},\"fields\": [\"PublicKey\"]}", organization)
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return ""
	}
	defer resultsIterator.Close()
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return ""
		}

		fmt.Println(string(queryResponse.Value))
		return string(queryResponse.Value)

	}
	return ""
}

//======================测试======================
func (t *SimpleChaincode) QuerySinglebyId(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	}
	fmt.Println(string(valAsbytes))
	return shim.Success(valAsbytes)
}

//查询银行-监管方
//参数："监管A"
func (t *SimpleChaincode) QueryMonitor2Bank(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke  QueryMonitor2Bank")
	argsLen := len(args)
	if argsLen < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	//var banks = args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"Bank2Monitor\",\"Monitor\":{\"$in\":[" + "\"" + args[0] + "\"" + "]}}}")

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//rule自定规则
//参数：id，过期时间小时数，投票规则数
func (t *SimpleChaincode) CreateRule(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke CreateRule")
	var err error

	result := t.VertifyArgs(3, args)
	if result != "" {
		return shim.Error(result)
	}

	Ruleid := args[0]
	DeadlineHourNum := args[1]
	AgreedNum := args[2]

	RuleBytes, err := stub.GetState(Ruleid)
	if err != nil {
		fmt.Println("Failed to get RuleBytes:" + Ruleid)
		return shim.Error("Failed to get RuleBytes: " + err.Error())
	} else if RuleBytes != nil {
		fmt.Println("This RuleBytes already exists: " + Ruleid)
		return shim.Error("This Nostro already exists: " + Ruleid)
	}
	objectType := "Rule"

	RuleStruct := &Rule{objectType, Ruleid, DeadlineHourNum, AgreedNum}
	RuleStructJSONasBytes, err := json.Marshal(RuleStruct)
	fmt.Println(RuleStructJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	//调用方法将数据写入账本（对于简单的POC，可以简便的使用clrbkcde写入，对后续的查询会方便许多）
	err = stub.PutState(Ruleid, RuleStructJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("Creat Rule is success")
	return shim.Success(nil)
}

//账本数据初始化
//参数，id， 本银行名，对方银行名， 货币代码，金额
func (t *SimpleChaincode) InitNoStro(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke InitNoStro")
	var err error

	result := t.VertifyArgs(5, args)
	if result != "" {
		return shim.Error(result)
	}

	actid := args[0]
	bkcode := args[1]
	clrbkcde := args[2]
	curcde := args[3]
	nostrobal := args[4]

	bal, err := strconv.Atoi(nostrobal)
	if err != nil {
		return shim.Error("5th argument nostrobal  must be number")
	}

	if bal < 0 {
		return shim.Error("5th argument nostrobal  must be > 0")
	}

	NostroBytes, err := stub.GetState(actid)
	if err != nil {
		fmt.Println("Failed to get Nostro:" + actid)
		return shim.Error("Failed to get Nostro: " + err.Error())
	} else if NostroBytes != nil {
		fmt.Println("This Nostro already exists: " + actid)
		return shim.Error("This Nostro already exists: " + actid)
	}
	objectType := NOSTROTYPE
	//actid = bkcode + clrbkcde + curcde
	nostro := &TdNoStro{objectType, actid, bkcode, clrbkcde, curcde, nostrobal}
	nostroJSONasBytes, err := json.Marshal(nostro)
	fmt.Println(nostroJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	//调用方法将数据写入账本（对于简单的POC，可以简便的使用clrbkcde写入，对后续的查询会方便许多）
	err = stub.PutState(actid, nostroJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("init is success")
	return shim.Success(nil)
}

//账本数据更新
//参数，id， 本银行名，对方银行名， 货币代码，金额
func (t *SimpleChaincode) UpdateNoStro(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke InitNoStro")
	var err error

	result := t.VertifyArgs(5, args)
	if result != "" {
		return shim.Error(result)
	}

	actid := args[0]
	bkcode := args[1]
	clrbkcde := args[2]
	curcde := args[3]
	nostrobals := args[4]

	bal, err := strconv.Atoi(nostrobals)
	if err != nil {
		return shim.Error("5th argument nostrobal  must be number")
	}

	if bal < 0 {
		return shim.Error("5th argument nostrobal  must be > 0")
	}

	//更改Nostro
	nostrobal := strconv.Itoa(bal)
	objectType := NOSTROTYPE
	nostro := &TdNoStro{objectType, actid, bkcode, clrbkcde, curcde, nostrobal}
	nostroJSONasBytes, err := json.Marshal(nostro)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("New nostro:" + string(nostroJSONasBytes))

	nostroId := actid
	err = stub.PutState(nostroId, nostroJSONasBytes)
	//判断是否成功写入账本中
	if err != nil {
		fmt.Println("Balance failed")
		return shim.Error(err.Error())
	}
	fmt.Println("Balance is completed")
	return shim.Success(nil)
}

//汇款交易
//参数：转出行，转出账户，转入行，转入账号，转账金额，货币代码，转账日期，转账时间，
func (t *SimpleChaincode) TxInstr(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("Invoke TxInstr")
	var err error

	result := t.VertifyArgs(10, args)
	if result != "" {
		return shim.Error(result)
	}

	actid := args[0]
	bkcode := args[1]
	actnofrom := args[2]
	clrbkcde := args[3]
	actnoto := args[4]
	txamt := args[5] //交易金额
	curcde := args[6]
	txndat := args[7]
	txntime := args[8]
	compst := "S"
	rtcde := ""
	secretkeyid := args[9]
	//判断如果交易存在，则需要重新生成
	exist, err1 := stub.GetState(actid)
	if err1 != nil {
		fmt.Println("failed to get tx info:" + actid)
		return shim.Error("failed to get tx info ")
	}

	if exist != nil {
		fmt.Println("id already exists, please change: " + actid)
		return shim.Error("id already exists, please change")
	}

	//定义一个object变量
	objectType := TXTYPE
	txinstr := &TxInstr{objectType, actid, bkcode, actnofrom, clrbkcde, actnoto, txamt, curcde, txndat, txntime, compst, rtcde, secretkeyid}
	txinstrJSONasBytes, err := json.Marshal(txinstr)
	if err != nil {
		fmt.Println("convert json error")
		return shim.Error(err.Error())
	}
	err = stub.PutState(actid, txinstrJSONasBytes)
	if err != nil {
		fmt.Println("PutState Error")
		return shim.Error(err.Error())
	}
	fmt.Println("Invoke completed")
	return shim.Success(nil)
}

//通过交易流水号查询单个交易信息
//参数：需要传入交流号
func (t *SimpleChaincode) QueryTransaction(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke  QueryTransaction")
	result := t.VertifyArgs(1, args)
	if result != "" {
		return shim.Error(result)
	}
	actid := args[0]
	Avalbytes, err := stub.GetState(actid)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + actid + "\"}"
		fmt.Println(jsonResp)
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + actid + "\"}"
		fmt.Println(jsonResp)
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"actid\":\"" + actid + "\",\"value\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

//查询NOSTRO数据，
//参数：银行号
func (t *SimpleChaincode) QueryTdNoStro(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke  QueryTdNoStro")
	var queryString string

	argsLen := len(args)
	if argsLen < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	//,\"BKCODE\":\"%s\"
	queryString = fmt.Sprintf("{\"selector\":{\"docType\":\"nostro\"")

	for i := 0; i < argsLen; i++ {
		if len(args[i]) <= 0 {
			return shim.Error(strconv.Itoa(i) + " argument must be a non-empty string")
		}
		if i == 0 {
			queryString = queryString + ",\"bkcode\":\"" + args[i] + "\""
		} else if i == 1 {
			queryString = queryString + ",\"clrbkcde\":\"" + args[i] + "\""
		} else if i == 2 {
			queryString = queryString + ",\"curcde\":\"" + args[i] + "\""
		}
	}
	queryString = queryString + "}}"
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//修改交易状态
//参数：字符数组，第一个参数：交易流水ID，第二个参数为修改后的状态值“S”或“R”
//此方法中查询都根据存储的key值进行查询，不需要创建组合键
func (t *SimpleChaincode) UpdateCompst(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke UpdateCompst")
	result := t.VertifyArgs(2, args)
	if result != "" {
		return shim.Error(result)
	}

	actid := args[0]
	compst := args[1]

	if compst != "S" && compst != "P" && compst != "R" {
		fmt.Println("wrong tx state ")
		return shim.Error("wrong tx state:" + compst)
	}

	var txinstrJson TxInstr
	var err error
	//查询账本中是否有对应ID的数据
	TxInstrBytes, err := stub.GetState(actid)
	if err != nil {
		fmt.Println("There is no such data in the books")
		return shim.Error("There is no such data in the books")
	}
	err = json.Unmarshal([]byte(TxInstrBytes), &txinstrJson)
	bkcode := txinstrJson.BKCODE
	actnofrom := txinstrJson.ACTNOFROM
	clrbkcde := txinstrJson.CLRBKCDE
	actnoto := txinstrJson.ACTNOTO
	txamt := txinstrJson.TXAMT
	curcde := txinstrJson.CURCDE
	txndat := txinstrJson.TXNDAT
	txntime := txinstrJson.TXNTIME
	rtcde := txinstrJson.RTCDE
	//publickeyid := txinstrJson.PUBLICKEYID
	secretkeyid := txinstrJson.SECRETKEYID
	objectType := TXTYPE
	txinstr := &TxInstr{objectType, actid, bkcode, actnofrom, clrbkcde, actnoto, txamt, curcde, txndat, txntime, compst, rtcde, secretkeyid}
	txinstrJSONasBytes, err := json.Marshal(txinstr)
	//更新TxInstr的数据
	fmt.Println("UpdateCompst, update tx 2:" + string(txinstrJSONasBytes))
	err = stub.PutState(actid, txinstrJSONasBytes)
	if err != nil {
		fmt.Println("UpdateCompst:TxInstr Write data failed")
		return shim.Error("TxInstr Write data failed")
	}
	t.triggerStatusUpdateEvent(stub, txinstr)
	return shim.Success(nil)
}

//用户汇入查询,用于CouchDb中
//参数：只需要ACTNOTO（转入账号）
func (t *SimpleChaincode) QueryTxInstrByActnoto(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("Invoke QueryTxInstrByActnoto")

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	if len(args[0]) <= 0 {
		return shim.Error(" argument must be a non-empty string")
	}

	actnoto := args[0] //转入账号

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"txinstr\",\"actnoto\":\"%s\"}}", actnoto)
	fmt.Println(queryString)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//用户汇出查询,用于CouchDb中
//参数：只需要ACTNOFROM（转出账号）
func (t *SimpleChaincode) QueryTxInstrByActnofrom(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke QueryTxInstrByActnofrom")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	if len(args[0]) <= 0 {
		return shim.Error(" argument must be a non-empty string")
	}
	actnofrom := args[0] //转出账号

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"txinstr\",\"actnofrom\":\"%s\"}}", actnofrom)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//用户条件查询,用于CouchDb中
//参数：需要actnofrom actnoto txndat
func (t *SimpleChaincode) UserQuery(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke UserQuery")
	var queryString = "{\"selector\":{\"docType\":\"txinstr\""

	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	if len(args[0]) > 0 {
		queryString += ",\"actnofrom\":\"" + args[0] + "\""
	}
	if len(args[1]) > 0 {
		queryString += ",\"actnoto\":\"" + args[1] + "\""
	}
	if len(args[2]) > 0 {
		queryString += ",\"txndat\":\"" + args[2] + "\""
	}
	queryString += "}}"

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//银行条件查询,用于CouchDb中
//参数：需要bkcode clrbkcde actnofrom actnoto txndat
func (t *SimpleChaincode) BankQuery(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke UserQuery")
	if len(args) < 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	var queryString = "{\"selector\":{\"docType\":\"txinstr\""

	if len(args[0]) > 0 {
		queryString += ",\"bkcode\":\"" + args[0] + "\""
	}
	if len(args[1]) > 0 {
		queryString += ",\"clrbkcde\":\"" + args[1] + "\""
	}
	if len(args[2]) > 0 {

		queryString += ",\"actnofrom\":\"" + args[2] + "\""
	}
	if len(args[3]) > 0 {
		queryString += ",\"actnoto\":\"" + args[3] + "\""
	}
	if len(args[4]) > 0 {
		queryString += ",\"txndat\":\"" + args[4] + "\""
	}

	queryString += "}}"

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

//银行汇入查询,用于CouchDb中
//参数：只需要CLRBKCDE（转入银行号）
func (t *SimpleChaincode) QueryTxInstrByClrbkcde(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke QueryTxInstrByClrbkcde")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	if len(args[0]) <= 0 {
		return shim.Error(" argument must be a non-empty string")
	}
	clrbkcde := args[0] //转入银行号

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"txinstr\",\"clrbkcde\":\"%s\"}}", clrbkcde)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//银行汇出查询,用于CouchDb中
//参数：只需要BKCODE（转出银行号）
func (t *SimpleChaincode) QueryTxInstrByBkcode(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke QueryTxInstrByClrbkcde")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	if len(args[0]) <= 0 {
		return shim.Error(" argument must be a non-empty string")
	}
	bkcode := args[0] //转出银行号

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"txinstr\",\"bkcode\":\"%s\"}}", bkcode)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//校验参数
func (t *SimpleChaincode) VertifyArgs(count int, args []string) string {
	argsLen := len(args)
	if argsLen < count {
		return "Incorrect number of arguments. Expecting " + strconv.Itoa(count)
	}

	for i := 0; i < argsLen; i++ {
		if len(args[i]) <= 0 {
			return strconv.Itoa(i) + " argument must be a non-empty string"
		}
	}

	return ""
}

//汇款汇入汇出的具体查询方法
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)
	//调用GetQueryResult接口queryString是底层状态数据库的查询语法，返回一个包含所需记录的迭代器
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	//定义一个字节输出流
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		//循环拿到插入到数据库中的key value
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
	//返回为byte[]数组，输出之后自动转为json
	return buffer.Bytes(), nil
}

func (t *SimpleChaincode) QueryByString(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke QueryByString")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	if len(args[0]) <= 0 {
		return shim.Error(" argument must be a non-empty string")
	}

	queryString := args[0] //转入账号

	fmt.Println(queryString)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//入口main
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

//=====================================================新增方法=========================================================
//获取证书（权限控制：1.相应的机构功能选择不一样 2.同一机构下的各个节点做不同的事情（待考虑））
func (t *SimpleChaincode) GetCertificate(stub shim.ChaincodeStubInterface, args []string) (organization string) {
	creatorByte, _ := stub.GetCreator()
	certStart := bytes.IndexAny(creatorByte, "-----")
	if certStart == -1 {
		fmt.Errorf("No certificate found")
	}
	certText := creatorByte[certStart:]
	bl, _ := pem.Decode(certText)
	if bl == nil {
		fmt.Errorf("Could not decode the PEM structure")
	}

	cert, err := x509.ParseCertificate(bl.Bytes)
	if err != nil {
		fmt.Errorf("ParseCertificate failed")
	}
	Organization := cert.Issuer.Organization
	organization2 := Organization[0]
	//uname := cert.Subject.CommonName
	fmt.Println(organization)
	//fmt.Println("Name:" + uname)

	//array := [2]string{organization, uname}
	return organization2
}

//公钥上链  把修改和 更新写在一起
//参数，id， 机构名称，机构证书名， 公钥，时间，状态  无用
func (t *SimpleChaincode) InitPublicKey3(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke InitPublicKey")
	var err error

	result := t.VertifyArgs(7, args)
	if result != "" {
		return shim.Error(result)
	}
	Pubid := args[0]
	Role := args[1]
	Dt := args[2]
	OrganizationCertName := args[3]
	Status := args[4] //通过传进来的状态判断：新增  Update修改
	PublicKey := args[5]
	OrganizationName := args[6]

	PublicKeyOld, err := stub.GetState(Pubid)
	//这里增加一个status的判断是为了判断
	if PublicKeyOld != nil && Status != "update" {
		fmt.Println("This publickey already exists: " + Pubid)
		return shim.Error("This publickey already exists: " + Pubid)
	}

	//主要是查询到Status为New的那条记录，修改其状态为”update“
	if Status == "update" { //修改之前的公钥状态，并插入这条公钥
		//这是每次只保证一条是正常的
		queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"PublicKey\",\"Status\":\"latest\",\"OrganizationName\":\"%s\"}}", OrganizationName)
		//换一下,直接通过getState().就可以取出。用多条查询是为了防止，功能又改变，到时候又改逻辑。。。
		//queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"PublicKey\",\"Status\":\"latest\",\"Pubid\":\"%s\"}}", Pubid)
		resultsIterator, err := stub.GetQueryResult(queryString)
		if err != nil {
			return shim.Error("This resultsIterator does't already exists1 ")
		}
		defer resultsIterator.Close()
		for resultsIterator.HasNext() {
			queryResponse, err := resultsIterator.Next()
			if err != nil {
				return shim.Error("This publickey does't already exists2 ")
			}

			//fmt.Println(string(queryResponse.Value))
			fmt.Println(string(queryResponse.Value))
			//将得到的Value解析为结构体用来操作
			UpdatePublicKey := new(PublicKeyStruct)
			err = json.Unmarshal(queryResponse.Value, UpdatePublicKey)
			UpdatePublicKey.PublicKey = PublicKey //args[5]
			UpdatePublicKey.Status = "update"
			//重新插入
			fmt.Println(UpdatePublicKey)
			UpdatePublicKeyJSONasBytes, err := json.Marshal(UpdatePublicKey)

			err = stub.PutState(queryResponse.Key, UpdatePublicKeyJSONasBytes)
			if err != nil {
				return shim.Error(err.Error())
			}

		}

	}
	//插入新增的
	objectType := "PublicKey"
	publickeyStruct2 := &PublicKeyStruct{objectType, Pubid, Role, Dt, OrganizationCertName, "latest", PublicKey, OrganizationName}
	publickeyJSONasBytes, err := json.Marshal(publickeyStruct2)
	fmt.Println(publickeyStruct2)
	fmt.Println(publickeyJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(Pubid, publickeyJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("init publickey is success")
	return shim.Success(nil)
}

//第二种办法InitPublicKey2==============================
type InitPublicKeyData struct {
	PublicKeyStruct []*PublicKeyStruct
}

//初始化Publickey数据
func (t *SimpleChaincode) InitPublicKey2(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	data := new(InitPublicKeyData)
	err := ParseJsonAndDecode(data, args)
	if err != nil {
		log.Println("Error occurred when parsing json")
		return shim.Error(err.Error())
	}
	for i := 0; i < len(data.PublicKeyStruct); i++ {
		InsertPublicKey(stub, data.PublicKeyStruct[i])
	}

	return shim.Success(nil)
}
func InsertPublicKey(stub shim.ChaincodeStubInterface, PublicKey *PublicKeyStruct) pb.Response {

	fmt.Println("=======insert=======")
	PublicKeyJSONasBytes, err := json.Marshal(PublicKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(PublicKey.Pubid, PublicKeyJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

//第三次修改
type UpdatePublicKeyAndSecData struct {
	Publickey       string
	UpdatePubAndSec []*UpdatePubAndSec
}
type UpdatePubAndSec struct {
	Id     string
	Secret string
}

//传入证书的名字
func (t *SimpleChaincode) InitPublicKey(stub shim.ChaincodeStubInterface, args []string, OrganizationCert string) pb.Response {
	fmt.Println("Invoke InitPublicKey")

	argsLen := len(args)
	if argsLen == 7 {
		result := t.VertifyArgs(7, args)
		if result != "" {
			return shim.Error(result)
		}
		Pubid := args[0]
		Role := args[1]
		Dt := args[2]
		OrganizationCertName := args[3]
		Status := args[4]
		PublicKey := args[5]
		OrganizationName := args[6]
		//第一步校验公钥存不存在不存在就新增===start===
		queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"PublicKey\",\"Status\":\"latest\",\"OrganizationCertName\":\"%s\"}}", OrganizationCert)
		resultsIterator, _ := stub.GetQueryResult(queryString)
		defer resultsIterator.Close()
		for resultsIterator.HasNext() {
			queryResponse, _ := resultsIterator.Next()
			getQueryResult := string(queryResponse.Value)
			if getQueryResult != "" {
				return shim.Error("公钥已经上链，只能修改")
			}
		}
		//================校验是不是有公钥=====end==
		//插入新增的
		objectType := "PublicKey"
		publickeyStruct2 := &PublicKeyStruct{objectType, Pubid, Role, Dt, OrganizationCertName, Status, PublicKey, OrganizationName}
		publickeyJSONasBytes, err := json.Marshal(publickeyStruct2)
		fmt.Println(publickeyStruct2)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = stub.PutState(Pubid, publickeyJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
		fmt.Println("init publickey is success")
		return shim.Success(nil)
	} else {
		data := new(UpdatePublicKeyAndSecData)
		err := ParseJsonAndDecode(data, args)
		if err != nil {
			log.Println("Error occurred when parsing json")
			return shim.Error(err.Error())
		}
		//角色
		var GavOrBank = ""
		//机构名字
		var NameOfOrg = ""
		//首先修改公钥start
		queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"PublicKey\",\"Status\":\"latest\",\"OrganizationCertName\":\"%s\"}}", OrganizationCert)
		resultsIterator, _ := stub.GetQueryResult(queryString)
		defer resultsIterator.Close()
		for resultsIterator.HasNext() {
			queryResponse, _ := resultsIterator.Next()
			fmt.Println(string(queryResponse.Value))
			//将得到的Value解析为结构体用来操作
			UpdatePublicKey := new(PublicKeyStruct)
			err = json.Unmarshal(queryResponse.Value, UpdatePublicKey)
			UpdatePublicKey.PublicKey = data.Publickey //args[5]
			GavOrBank = UpdatePublicKey.Role
			//====修改的时间===========
			now := time.Now() //当前的时间
			timeNow := now.Format("2006-01-02 15:04:05")
			fmt.Println("now(time format):", now)
			fmt.Println("tNow(string format):", timeNow)
			UpdatePublicKey.Dt = timeNow
			//======修改时间完成======
			NameOfOrg = UpdatePublicKey.OrganizationName
			//重新插入
			fmt.Println(UpdatePublicKey)
			UpdatePublicKeyJSONasBytes, err := json.Marshal(UpdatePublicKey)

			err = stub.PutState(queryResponse.Key, UpdatePublicKeyJSONasBytes)
			if err != nil {
				return shim.Error(err.Error())
			}

		}
		//修改公钥完毕
		//修改密钥开始
		for i := 0; i < len(data.UpdatePubAndSec); i++ {

			updateSecretData(stub, data.UpdatePubAndSec[i], GavOrBank, NameOfOrg)
		}

		return shim.Success(nil)
	}

}

//公钥失效后重新插入
func updateSecretData(stub shim.ChaincodeStubInterface, updateSec *UpdatePubAndSec, GavOrBank string, Name string) pb.Response {
	var jsonResp string
	fmt.Println("=======更改原来的密钥调用方法updateSecretData=======")

	valAsbytes, err := stub.GetState(updateSec.Id) //获取json
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + updateSec.Id + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"SecretKey does not exist: " + updateSec.Id + "\"}"
		return shim.Error(jsonResp)
	}
	jsOld, _ := NewJson([]byte(valAsbytes))
	jsOld.SetPath([]string{"Secret", GavOrBank, Name, "secret"}, updateSec.Secret) //增加监管机构
	jsOld2 := jsOld.MustMap(map[string]interface{}{"found": false})                //数据转化一下
	SecretKeyOldJSONasBytes, _ := json.Marshal(jsOld2)
	fmt.Println(string(SecretKeyOldJSONasBytes))
	err = stub.PutState(updateSec.Id, SecretKeyOldJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	jsonResp = "{\"success\":\"UpdateSecretKey Id: " + updateSec.Id + "\"}"
	return shim.Success(nil)
}

//================================================================================
//获取对应的角色Role和机构名
func GetRoleAndOrgName(stub shim.ChaincodeStubInterface, organization string) (string, string) {
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"PublicKey\",\"Status\":\"latest\",\"OrganizationCertName\":\"%s\"}}", organization)
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return "", ""
	}
	defer resultsIterator.Close()
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return "", ""
		}

		UpdatePublicKey := new(PublicKeyStruct)
		err = json.Unmarshal(queryResponse.Value, UpdatePublicKey)

		return UpdatePublicKey.Role, UpdatePublicKey.OrganizationName

	}
	return "", ""
}

//当生成对称密钥的时候插入区块链
func (t *SimpleChaincode) IncreaseSecretKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//直接插入值  获得传进来的值
	//SecretKeyJson := args[0] //直接传进来的参数cli无法测试

	SecretKeyJson, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		log.Println("ParseJson base64 decode error.")
		return shim.Error("ParseJson base64 decode error.")
	}
	fmt.Println("data after decode fmt:" + string(SecretKeyJson[:]))
	js, _ := NewJson([]byte(SecretKeyJson))
	UUID, _ := js.Get("id").String()
	jsNew := js.MustMap(map[string]interface{}{"found": false}) //直接存json数据拿到的时候会很奇怪
	SecretKeyJsonasBytes, _ := json.Marshal(jsNew)
	//SecretKeyJsonasBytes, _ := json.Marshal(SecretKeyJson)
	err = stub.PutState(UUID, SecretKeyJsonasBytes)
	if err != nil {
		return shim.Error("IncreaseSecretKey Error")
	}

	return shim.Success(nil)
}

//修改加密私钥的有效性
func (t *SimpleChaincode) UpdateSecretKey2(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var uuid, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	uuid = args[0]
	valAsbytes, err := stub.GetState(uuid) //
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + uuid + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"SecretKey does not exist: " + uuid + "\"}"
		return shim.Error(jsonResp)
	}
	jsOld, _ := NewJson([]byte(valAsbytes))
	jsOld.Set("state", "N")
	jsOld2 := jsOld.MustMap(map[string]interface{}{"found": false}) //数据转化一下
	SecretKeyOldJSONasBytes, _ := json.Marshal(jsOld2)
	fmt.Println(string(SecretKeyOldJSONasBytes))
	err = stub.PutState(uuid, SecretKeyOldJSONasBytes)
	if err != nil {
		return shim.Error("IncreaseSecretKey Error")
	}
	jsonResp = "{\"success\":\"UpdateSecretKey Id: " + uuid + "\"}"
	return shim.Success(nil)
}

//存储bank和监管机构的关系
func (t *SimpleChaincode) BankAndMonitor(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//直接插入值  获得传进来的值
	//SecretKeyJson := args[0] //直接传进来的参数cli无法测试

	SecretKeyJson, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		log.Println("ParseJson base64 decode error.")
		return shim.Error("ParseJson base64 decode error.")
	}
	fmt.Println("data after decode fmt:" + string(SecretKeyJson[:]))
	js, _ := NewJson([]byte(SecretKeyJson))
	BankAndMonitorId, _ := js.Get("Bank").String()
	jsNew := js.MustMap(map[string]interface{}{"found": false}) //直接存json数据拿到的时候会很奇怪
	BankAndMonitorJsonasBytes, _ := json.Marshal(jsNew)
	//SecretKeyJsonasBytes, _ := json.Marshal(SecretKeyJson)
	err = stub.PutState(BankAndMonitorId, BankAndMonitorJsonasBytes)
	if err != nil {
		return shim.Error("IncreaseSecretKey Error")
	}

	return shim.Success(nil)
}

//第三次修改
type AddMonitorData struct {
	AddMonitorName  string
	AddPubid        string
	UpdatePubAndSec []*UpdatePubAndSec
}

//加入新的监管机构后怎么解决问题
//这个等待开发===传进来的是所有需要监管的银行数据如下(bankA,bankB,bankC,bankD......)
//参数：需要监管的银行，机构名，公钥id,公钥加密的数据
func (t *SimpleChaincode) AddOrganizationPublicKey2Tx3(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	argsLen := len(args)
	//var jsonResp string
	if argsLen < 1 {

		return shim.Error("Error: Incorrect number of arguments. Expecting")
	}
	for i := 0; i < argsLen; i++ {
		if len(args[i]) <= 0 {

			return shim.Error("Error:  argument must be a non-empty string")
		}

	}

	//1.需要修改密钥。2.需要修改银行受那家监管机构监管。3.需要验证防止重复插入

	data := new(AddMonitorData)
	err := ParseJsonAndDecode(data, args)
	if err != nil {
		log.Println("Error occurred when parsing json")
		return shim.Error(err.Error())
	}
	//监管那些机构
	//BeMonitoredBnaks := ""
	//校验这个机构是否投票同意
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"CreateVote\",\"Opinion\":\"%s\",\"OrganizationName\":\"%s\"}}", "S", data.AddMonitorName)
	resultsIterator, _ := stub.GetQueryResult(queryString)
	defer resultsIterator.Close()
	for resultsIterator.HasNext() {
		queryResponse, _ := resultsIterator.Next()
		getQueryResult := string(queryResponse.Value)
		if getQueryResult == "" {
			return shim.Error("投票没成功的机构")
		}
		//如果成功得到它监管的组织
		VoteStruct := new(Vote)
		err := json.Unmarshal(queryResponse.Value, VoteStruct)
		if err != nil {
			return shim.Error(err.Error())
		}
		//BeMonitoredBnaks = VoteStruct.BanktoMonitor

	}

	//先修改密钥
	for i := 0; i < len(data.UpdatePubAndSec); i++ {
		addMonitorSecretData(stub, data.UpdatePubAndSec[i], data.AddMonitorName, data.AddPubid)
	}

	//增加银行的与监管机构的关系
	//	BankArray := strings.Split(BeMonitoredBnaks, ",") //得到传进来的参数Split
	//	for i := 0; i < len(BankArray); i++ {
	//		valAsbytes, err := stub.GetState(BankArray[i]) //获取json
	//		if err != nil {
	//			jsonResp = "{\"Error\":\"Failed to get state for " + BankArray[i] + "\"}"
	//			return shim.Error(jsonResp)
	//		} else if valAsbytes == nil {
	//			jsonResp = "{\"Error\":\"SecretKey does not exist: " + BankArray[i] + "\"}"
	//			return shim.Error(jsonResp)
	//		}
	//		jsces2, _ := NewJson([]byte(valAsbytes))
	//		arryMonitor, _ := jsces2.Get("Monitor").Array()

	//		fmt.Println(arryMonitor)
	//		arryMonitor = append(arryMonitor, data.AddMonitorName)
	//		fmt.Println(arryMonitor)
	//		jsces2.Set("Monitor", arryMonitor)
	//		jsces3 := jsces2.MustMap(map[string]interface{}{"found": false}) //数据转化一下
	//		JSONasBytes, _ := json.Marshal(jsces3)
	//		fmt.Println(string(JSONasBytes))
	//		err = stub.PutState(BankArray[i], JSONasBytes)
	//		if err != nil {
	//			return shim.Error(err.Error())
	//		}

	//	}
	return shim.Success(nil)
}

//添加监管机构的公钥和加密密钥
func addMonitorSecretData(stub shim.ChaincodeStubInterface, updateSec *UpdatePubAndSec, Name string, pubid string) pb.Response {
	var jsonResp string
	fmt.Println("=======更改原来的密钥调用方法addMonitorSecretData=======")

	valAsbytes, err := stub.GetState(updateSec.Id) //获取json
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + updateSec.Id + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"SecretKey does not exist: " + updateSec.Id + "\"}"
		return shim.Error(jsonResp)
	}
	jsOld, _ := NewJson([]byte(valAsbytes))
	jsOld.SetPath([]string{"Secret", "Monitor", Name, "pubid"}, pubid) //增加监管机构
	jsOld.SetPath([]string{"Secret", "Monitor", Name, "secret"}, updateSec.Secret)
	jsOld2 := jsOld.MustMap(map[string]interface{}{"found": false}) //数据转化一下
	SecretKeyOldJSONasBytes, _ := json.Marshal(jsOld2)
	fmt.Println(string(SecretKeyOldJSONasBytes))
	err = stub.PutState(updateSec.Id, SecretKeyOldJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	jsonResp = "{\"success\":\"addMonitorSecretData Id: " + updateSec.Id + "\"}"
	return shim.Success(nil)
}

//触发相应的操作
func (t *SimpleChaincode) triggerStatusUpdateEvent(stub shim.ChaincodeStubInterface, txInstr *TxInstr) pb.Response {

	jsonResp, _ := json.Marshal(txInstr)
	fmt.Println("Trigger a event to the app")
	fmt.Println(string(jsonResp))
	//Trigger a event to the app
	err := stub.SetEvent("TxInstr_Update", jsonResp)
	if err != nil {
		return shim.Error("Error: failed to settriggerStatusUpdateEvent event")
	}

	return shim.Success([]byte(fmt.Sprintf("Triggered event to update txInstr status: %s", txInstr.INSTRID)))
}

//发起投票都是一样的
//c参数：id,监管机构的名字，投票机构，需要投票的机构，发起投票的机构，状态C:发起机构创建 规则id
func (t *SimpleChaincode) CreateVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke CreateVote")
	var err error
	Voteid := args[0]
	OrganizationName := args[1]
	//VoteOrganization := args[2]//?
	//DeadlineDate := args[3]
	//	VoteDate ：=args[4]
	BanktoMonitor := args[2]
	InitiateOrganization := args[3]
	Opinion := args[4]
	RuleID := args[5]
	//校验公钥上链--start--
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"PublicKey\",\"Status\":\"latest\",\"OrganizationName\":\"%s\"}}", OrganizationName)
	fmt.Println(queryString)
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error("公钥没上链: " + OrganizationName)
	}
	defer resultsIterator.Close()
	var numP = 0
	for resultsIterator.HasNext() {

		queryResponse, _ := resultsIterator.Next()

		getQueryResult := string(queryResponse.Value)
		fmt.Println("====================")
		numP++

		fmt.Println(strconv.Itoa(numP))
		fmt.Println(getQueryResult)
		if getQueryResult == "" {
			return shim.Error("公钥没上链: " + OrganizationName)
		}

	}

	if numP == 0 {
		return shim.Error("公钥没上链: " + OrganizationName)
	}
	//校验公钥上链--end--
	//是否需要校验同一组织发起成为监管机构多次？待处理
	if Opinion == "C" {
		VoteBytes, err := stub.GetState(Voteid)
		if err != nil {
			fmt.Println("Failed to get VoteBytes:" + Voteid)
			return shim.Error("Failed to get VoteBytes: " + err.Error())
		} else if VoteBytes != nil {
			fmt.Println("This VoteBytes already exists: " + Voteid)
			return shim.Error("This VoteBytes already exists: " + Voteid)
		}
		//时间处理
		ruleBytesAsJSON, _ := stub.GetState(RuleID)
		var Rule2 Rule
		json.Unmarshal(ruleBytesAsJSON, &Rule2)
		//时间处理start
		now := time.Now() //当前的时间
		fmt.Println(Rule2)
		var DeadlineHourNumFromRule = Rule2.DeadlineHourNum
		fmt.Println("延长时间规则：" + DeadlineHourNumFromRule)
		addDeadlineTime, _ := time.ParseDuration(DeadlineHourNumFromRule)
		addDeadlineTime2 := now.Add(addDeadlineTime)
		DeadlineDate := addDeadlineTime2.Format("2006-01-02 15:04:05")
		fmt.Println(now)
		fmt.Println("发起投票过期时间：" + DeadlineDate)
		//时间处理end
		objectType := "CreateVote"
		vote := &Vote{objectType, Voteid, OrganizationName, "", DeadlineDate, "", BanktoMonitor, InitiateOrganization, Opinion, RuleID}
		voteJSONasBytes, err := json.Marshal(vote)
		fmt.Println(voteJSONasBytes)
		if err != nil {
			return shim.Error("voteJSONasBytes Marshal")
		}
		err = stub.PutState(Voteid, voteJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
		OrgAndBankArray := strings.Split(BanktoMonitor, ",") //得到传进来的参数Split
		//将需要投票的银行写进入 Opinion：Wait W
		OrgAndBankArrayLen := len(OrgAndBankArray)
		for i := 0; i < OrgAndBankArrayLen; i++ {
			objectType2 := "Vote"
			VoteOrganization2 := OrgAndBankArray[i]
			Voteid2 := Voteid + ":" + VoteOrganization2
			vote2 := &Vote{objectType2, Voteid2, OrganizationName, VoteOrganization2, DeadlineDate, "", BanktoMonitor, InitiateOrganization, "W", RuleID}
			fmt.Println("投票的银行：===" + VoteOrganization2)
			fmt.Println(vote2)
			vote2JSONasBytes, err := json.Marshal(vote2)
			//fmt.Println(vote2JSONasBytes)
			if err != nil {
				return shim.Error("vote2JSONasBytes Marshal")
			}
			err = stub.PutState(Voteid2, vote2JSONasBytes)
			if err != nil {
				return shim.Error(err.Error())
			}

		}
		t.triggerVoteEvent(stub, vote)
	}

	fmt.Println("vote is success")
	return shim.Success(nil)
}

//投票
//参数：投票的银行:通过证书获得，Opinion：N/Y ，截至日期,规则id
func (t *SimpleChaincode) Vote(stub shim.ChaincodeStubInterface, args []string, orgOrBankName string) pb.Response {
	fmt.Println("Invoke Vote")
	//var err error
	Voteid := args[0]                          //触发后你保存到数据库的或者你查询出来的发起投票的id
	zuheVoteid := Voteid + ":" + orgOrBankName //该ID
	//DeadlineDate :=args[1] //截至日期
	Opinion := args[1] //给的建议选择N/Y
	RuleID := args[2]  //规则Id
	VoteBytes, _ := stub.GetState(zuheVoteid)
	var voteStruct Vote
	json.Unmarshal(VoteBytes, &voteStruct)
	now := time.Now() //现在时间
	timeNow := now.Format("2006-01-02 15:04:05")
	voteStruct.VoteDate = timeNow                                                       //投票时间
	deadlineDateFormat, _ := time.Parse("2006-01-02 15:04:05", voteStruct.DeadlineDate) //时间形态
	trueOrFalse := deadlineDateFormat.After(now)

	//检验是不是重复投票
	ruleDoubleBytesAsJSON, _ := stub.GetState(zuheVoteid)
	var voteDouble Vote
	json.Unmarshal(ruleDoubleBytesAsJSON, &voteDouble)
	if voteDouble.Opinion == Opinion {
		return shim.Error("同一投票结果请不要投多次")
	}
	if trueOrFalse == true {
		fmt.Println("deadlineDateFormat在now之后，还没过期!")
		voteStruct.Opinion = Opinion
		voteJSONasBytes, err := json.Marshal(voteStruct)
		fmt.Println(voteJSONasBytes)
		if err != nil {
			return shim.Error("voteJSONasBytes Marshal")
		}
		err = stub.PutState(zuheVoteid, voteJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
		if Opinion == "Y" {
			num := t.QueryVoteNumber(stub, voteStruct.OrganizationName)
			ruleBytesAsJSON, _ := stub.GetState(RuleID)
			//对规则的同意数目进行判断
			var Rule2 Rule
			json.Unmarshal(ruleBytesAsJSON, &Rule2)
			ruleNnm, _ := strconv.Atoi(Rule2.AgreedNum)
			//对创建的createVote的状态进行判断
			VoteUpdateBytesAsJSON, _ := stub.GetState(Voteid)
			var voteResult Vote
			json.Unmarshal(VoteUpdateBytesAsJSON, &voteResult)
			if ruleNnm <= num+1 && voteResult.Opinion != "S" {

				//重新插入
				voteResult.Opinion = "S"
				voteResultJSONasBytes, err := json.Marshal(voteResult)
				fmt.Println(voteResultJSONasBytes)
				if err != nil {
					return shim.Error(err.Error())
				}

				err = stub.PutState(Voteid, voteResultJSONasBytes)
				if err != nil {
					return shim.Error(err.Error())
				}
				var jsonResp string
				//增加银行的与监管机构的关系start
				BankArray := strings.Split(voteResult.BanktoMonitor, ",") //得到传进来的参数Split
				for i := 0; i < len(BankArray); i++ {
					valAsbytes, err := stub.GetState(BankArray[i]) //获取json
					if err != nil {
						jsonResp = "{\"Error\":\"Failed to get state for " + BankArray[i] + "\"}"
						return shim.Error(jsonResp)
					} else if valAsbytes == nil {
						jsonResp = "{\"Error\":\"BankAndMonitor does not exist: " + BankArray[i] + "\"}"
						return shim.Error(jsonResp)
					}
					jsces2, _ := NewJson([]byte(valAsbytes))
					arryMonitor, _ := jsces2.Get("Monitor").Array()

					fmt.Println(arryMonitor)
					arryMonitor = append(arryMonitor, voteResult.OrganizationName)
					fmt.Println(arryMonitor)
					jsces2.Set("Monitor", arryMonitor)
					jsces3 := jsces2.MustMap(map[string]interface{}{"found": false}) //数据转化一下
					JSONasBytes, _ := json.Marshal(jsces3)
					fmt.Println(string(JSONasBytes))
					err = stub.PutState(BankArray[i], JSONasBytes)
					if err != nil {
						return shim.Error(err.Error())
					}

				}
				//增加银行的与监管机构的关系end

				t.triggerVoteSucessEvent(stub, voteResult.OrganizationName, voteResult.BanktoMonitor)
			}
		}

	} else {
		fmt.Println("deadlineDateFormat在now之前，过期啦!")
		//		VoteUpdateBytesAsJSON, _ := stub.GetState(Voteid)
		//		var voteResult Vote
		//		json.Unmarshal(VoteUpdateBytesAsJSON, &voteResult)
		//				//重新插入
		//		voteResult.Opinion="F"
		//		voteResultJSONasBytes, err := json.Marshal(voteResult)
		//		fmt.Println(voteResultJSONasBytes)
		//		if err != nil {
		//			return shim.Error(err.Error())
		//			}

		//		err = stub.PutState(Voteid, voteResultJSONasBytes)
		//		if err != nil {
		//			return shim.Error(err.Error())
		//			}
		return shim.Error("过期了")
	}

	fmt.Println("vote is success")
	return shim.Success(nil)
}

//投票成功返回公钥
func (t *SimpleChaincode) triggerVoteSucessEvent(stub shim.ChaincodeStubInterface, name string, BanktoMonitor string) pb.Response {

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"PublicKey\",\"Status\":\"latest\",\"OrganizationName\":\"%s\"}}", name)
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error("公钥没上链: " + name)
	}
	defer resultsIterator.Close()

	//==========================================
	//定义一个字节输出流
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		//循环拿到插入到数据库中的key value
		queryResponse, _ := resultsIterator.Next()

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"BanktoMonitor\":")
		buffer.WriteString("\"")
		buffer.WriteString(BanktoMonitor)
		buffer.WriteString("\"")

		buffer.WriteString(", \"PublicKey\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	err = stub.SetEvent("Vote_Success", buffer.Bytes())
	if err != nil {
		return shim.Error("Error: failed to triggerVoteSucessEvent event")
	}

	//==========================================

	return shim.Success([]byte(fmt.Sprintf("Triggered event to update secret, Name Of Org: %s", name)))
}

//触发相应的操作
func (t *SimpleChaincode) triggerVoteEvent(stub shim.ChaincodeStubInterface, vote *Vote) pb.Response {

	jsonResp, _ := json.Marshal(vote)
	fmt.Println("Trigger a event to the app")
	fmt.Println(string(jsonResp))
	//Trigger a event to the app
	err := stub.SetEvent("TxInstr_Update", jsonResp)
	if err != nil {
		return shim.Error("Error: failed to settriggerStatusUpdateEvent event")
	}

	return shim.Success([]byte(fmt.Sprintf("Triggered event to update txInstr status: %s", vote.Voteid)))
}

//获取投票数目
func (t *SimpleChaincode) QueryVoteNumber(stub shim.ChaincodeStubInterface, args string) int {
	fmt.Println("Invoke  QueryVoteNumber")

	OrganizationName := args
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"Vote\",\"Opinion\":\"Y\",\"OrganizationName\":\"%s\"}}", OrganizationName)
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		fmt.Println("Error: fail to execute `GetQueryResult`")
		return 0
	}
	defer resultsIterator.Close()
	voteNumber := 0
	for resultsIterator.HasNext() {
		resultsIterator.Next()

		voteNumber++
	}
	jsonResp := "{\"voteNumber\":\"" + strconv.Itoa(voteNumber) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)

	//return shim.Success(jsonResp)
	return voteNumber
}

//投票同意的具体数据
//参数：OrganizationName/All
func (t *SimpleChaincode) QueryVotedata(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke  QueryVote")
	result := t.VertifyArgs(2, args)
	if result != "" {
		return shim.Error(result)
	}
	OrganizationName := args[0]
	queryNumber := args[1]
	if queryNumber == "all" {
		queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"vote\",\"OrganizationName\":\"%s\"}}", OrganizationName)
		queryResults, err := getQueryResultForQueryString(stub, queryString)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(queryResults)
	} else {
		queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"vote\",\"Opinion\":\"Y\",\"OrganizationName\":\"%s\"}}", OrganizationName)
		queryResults, err := getQueryResultForQueryString(stub, queryString)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(queryResults)
	}
}

//查询公钥
//参数：机构名
func (t *SimpleChaincode) QueryPublicKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke  QueryPublicKey")
	result := t.VertifyArgs(1, args)
	if result != "" {
		return shim.Error(result)
	}
	OrganizationName := args[0]
	//queryNumber := args[1]

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"PublicKey\",\"Status\":\"latest\",\"OrganizationName\":\"%s\"}}", OrganizationName)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//查询加密密钥
//参数：banks
func (t *SimpleChaincode) QuerySecretKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke  QuerySecretKey")
	argsLen := len(args)
	if argsLen < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	//var banks = args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"SecretKey\",\"state\":\"Y\",\"$and\":[")

	for i := 0; i < argsLen; i++ {

		if len(args[i]) <= 0 {
			return shim.Error(strconv.Itoa(i) + " argument must be a non-empty string")
		}

		if i == 0 {
			queryString = queryString + "{\"include\":{\"$in\":[" + "\"" + args[0] + "\"" + "]}}"
		} else {
			queryString = queryString + ",{\"include\":{\"$in\":[" + "\"" + args[i] + "\"" + "]}}"
		}
	}
	queryString = queryString + "]}}"
	//queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"SecretKey",\"state\":\"Y\",\"banks":\"%s\"}}", banks)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//查询加密密钥
//参数：include
func (t *SimpleChaincode) QuerySecretKey2(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke  QueryPublicKey")
	result := t.VertifyArgs(1, args)
	if result != "" {
		return shim.Error(result)
	}
	BankABankBName := args[0]
	BankArray := strings.Split(BankABankBName, ",") //得到传进来的参数Split
	BankBBankAName := BankArray[1] + "," + BankArray[0]
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"SecretKey\",\"state\":\"Y\",\"$or\":[{\"include\":\"%s\"},{\"include\":\"%s\"}]}}", BankABankBName, BankBBankAName)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//查询所有的密钥
func (t *SimpleChaincode) QuerySecretKeyAll(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke  QuerySecretKeyAll")
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"SecretKey\"}}")
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//查询投票机构名
func (t *SimpleChaincode) QueryBanks(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Invoke  QueryBanks")
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"PublicKey\",\"Status\":\"latest\",\"Role\":\"Bank\"}}")
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//查询对应银行的密钥
func (t *SimpleChaincode) QuerySecretKey3(stub shim.ChaincodeStubInterface, orgName string) pb.Response {
	fmt.Println("Invoke  QuerySecretKey3")
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"SecretKey\",\"include\":{\"$in\":[" + "\"" + orgName + "\"" + "]}}}")

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//===========================================以下json的工具===========================================
type SuperObject interface{}

func ParseJsonAndDecode(data SuperObject, args []string) error {
	fmt.Println(data)
	//base64解码
	arg, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		log.Println("ParseJson base64 decode error.")
		return err
	}
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

//==============================工具===================================
type Json struct {
	data interface{}
}

// NewJson returns a pointer to a new `Json` object
// after unmarshaling `body` bytes
func NewJson(body []byte) (*Json, error) {
	j := new(Json)
	err := j.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return j, nil
}
func (j *Json) Get(key string) *Json {
	m, err := j.Map()
	if err == nil {
		if val, ok := m[key]; ok {
			return &Json{val}
		}
	}
	return &Json{nil}
}

func (j *Json) UnmarshalJSON(p []byte) error {
	return json.Unmarshal(p, &j.data)
}

// Map type asserts to `map`
func (j *Json) Map() (map[string]interface{}, error) {
	if m, ok := (j.data).(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

func New() *Json {
	return &Json{
		data: make(map[string]interface{}),
	}
}

// GetIndex returns a pointer to a new `Json` object
// for `index` in its `array` representation
//
// this is the analog to Get when accessing elements of
// a json array instead of a json object:
//    js.Get("top_level").Get("array").GetIndex(1).Get("key").Int()
func (j *Json) GetIndex(index int) *Json {
	a, err := j.Array()
	if err == nil {
		if len(a) > index {
			return &Json{a[index]}
		}
	}
	return &Json{nil}
}

// Array type asserts to an `array`
func (j *Json) Array() ([]interface{}, error) {
	if a, ok := (j.data).([]interface{}); ok {
		return a, nil
	}
	return nil, errors.New("type assertion to []interface{} failed")
}

// String type asserts to `string`
func (j *Json) String() (string, error) {
	if s, ok := (j.data).(string); ok {
		return s, nil
	}
	return "", errors.New("type assertion to string failed")
}

// Set modifies `Json` map by `key` and `value`
// Useful for changing single key/value in a `Json` object easily.
func (j *Json) Set(key string, val interface{}) {
	m, err := j.Map()
	if err != nil {
		return
	}
	m[key] = val
}

// SetPath modifies `Json`, recursively checking/creating map keys for the supplied path,
// and then finally writing in the value
func (j *Json) SetPath(branch []string, val interface{}) {
	if len(branch) == 0 {
		j.data = val
		return
	}

	// in order to insert our branch, we need map[string]interface{}
	if _, ok := (j.data).(map[string]interface{}); !ok {
		// have to replace with something suitable
		j.data = make(map[string]interface{})
	}
	curr := j.data.(map[string]interface{})

	for i := 0; i < len(branch)-1; i++ {
		b := branch[i]
		// key exists?
		if _, ok := curr[b]; !ok {
			n := make(map[string]interface{})
			curr[b] = n
			curr = n
			continue
		}

		// make sure the value is the right sort of thing
		if _, ok := curr[b].(map[string]interface{}); !ok {
			// have to replace with something suitable
			n := make(map[string]interface{})
			curr[b] = n
		}

		curr = curr[b].(map[string]interface{})
	}

	// add remaining k/v
	curr[branch[len(branch)-1]] = val
}

func (j *Json) GetPath(branch ...string) *Json {
	jin := j
	for _, p := range branch {
		jin = jin.Get(p)
	}
	return jin
}

func (j *Json) MustMap(args ...map[string]interface{}) map[string]interface{} {
	var def map[string]interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustMap() received too many arguments %d", len(args))
	}

	a, err := j.Map()
	if err == nil {
		return a
	}

	return def
}
