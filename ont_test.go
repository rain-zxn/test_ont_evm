package awesomeProject4

import (
	"awesomeProject/awesomeProject4/chainsdk"
	"awesomeProject/awesomeProject4/eccm_abi"
	"awesomeProject/awesomeProject4/msg"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	ontSDK "github.com/ontio/ontology-go-sdk"
	ontcommon "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/merkle"
	ontocccom "github.com/ontio/ontology/smartcontract/service/native/cross_chain/common"
	ontSDK1Client "github.com/rain-zxn/ontology-go-sdk/client"
	ontnode "github.com/polynetwork/bridge-common/chains/ont"
	polynode "github.com/polynetwork/bridge-common/chains/poly"
	pcom "github.com/polynetwork/poly/common"
	pplycommon "github.com/polynetwork/poly/native/service/cross_chain_manager/common"

	bc "github.com/ontio/ontology/http/base/common"

	"strings"
	"testing"
)

type EthereumSdk struct {
	rpcClient *rpc.Client
	rawClient *ethclient.Client
	url       string
}

func NewEthereumSdk(url string) (*EthereumSdk, error) {
	rpcClient, err1 := rpc.Dial(url)
	rawClient, err2 := ethclient.Dial(url)
	if rpcClient == nil || err1 != nil || rawClient == nil || err2 != nil {
		return nil, fmt.Errorf("ethereum node is not working!, err1: %v, err2: %v", err1, err2)
	}
	return &EthereumSdk{
		rpcClient: rpcClient,
		rawClient: rawClient,
		url:       url,
	}, nil
}

func HexStringReverse1(value string) string {
	aa, _ := hex.DecodeString(value)
	bb := HexReverse(aa)
	return hex.EncodeToString(bb)
}
func HexReverse(arr []byte) []byte {
	l := len(arr)
	x := make([]byte, 0)
	for i := l - 1; i >= 0; i-- {
		x = append(x, arr[i])
	}
	return x
}


func scan() (*msg.Tx){
	sdk := ontSDK.NewOntologySdk()
	client := sdk.NewRpcClient()
	client.SetAddress("http://43.128.242.133:20336")

	//ethsdk,err:=NewEthereumSdk("http://43.128.242.133:20339")
	//if err != nil {
	//	fmt.Println("NewEthereumSdk err:", err)
	//	return
	//}
	height := 11228
	events, err := sdk.GetSmartContractEventByBlock(uint32(height))
	if err != nil {
		fmt.Println("GetSmartContractEventByBlock err:", err)
		return nil
	}
	eccmAddr := "34d4a23A1FC0C694f0D74DDAf9D8d564cfE2D430"
	eccmReversed := HexStringReverse1(eccmAddr)
	fmt.Println(eccmReversed)
	flag := 0
	for _, event0 := range events {
		fmt.Println("event hash:",event0.TxHash)
		for _, notify := range event0.Notify {
			if notify.ContractAddress == eccmReversed {
				flag++
				states, ok := notify.States.(string)
				if !ok {
					fmt.Println("event info states is not string")
					continue
				}
				var data []byte
				data, err = hexutil.Decode(states)
				if err != nil {
					err = fmt.Errorf("decoding states err:%v", err)
					return nil
				}
				source := pcom.NewZeroCopySource(data)
				var storageLog StorageLog
				err = storageLog.Deserialization(source)
				if err != nil {
					return nil
				}

				var parsed abi.ABI
				parsed, err = abi.JSON(strings.NewReader(eccm_abi.EthCrossChainManagerABI))
				if err != nil {
					return nil
				}
				var event eccm_abi.EthCrossChainManagerCrossChainEvent
				err = parsed.UnpackIntoInterface(&event, "CrossChainEvent", storageLog.Data)
				if err != nil {
					return nil
				}
				fmt.Println("CrossChainEvent", event)


				tx := &msg.Tx{
					TxType:     msg.SRC,
					TxId:       msg.EncodeTxId(event.TxId),
					SrcHash:    HexStringReverse1(event0.TxHash),
					DstChainId: event.ToChainId,
					SrcHeight:  uint64(height),
					SrcParam:   hex.EncodeToString(event.Rawdata),
					SrcChainId: 5555,
					SrcProxy:   event.ProxyOrAssetContract.String(),
					DstProxy:   common.BytesToAddress(event.ToContract).String(),
				}
				jsontx,_:=json.Marshal(tx)
				fmt.Println(string(jsontx))
				return tx
			}
		}
	}
	return nil
}

type StorageLog struct {
	Address common.Address
	Topics  []common.Hash
	Data    []byte
}

func (self *StorageLog) Serialization(sink *pcom.ZeroCopySink) {
	sink.WriteAddress(pcom.Address(self.Address))
	sink.WriteUint32(uint32(len(self.Topics)))
	for _, t := range self.Topics {
		sink.WriteHash(pcom.Uint256(t))
	}
	sink.WriteVarBytes(self.Data)
}

func (self *StorageLog) Deserialization(source *pcom.ZeroCopySource) error {
	address, _ := source.NextAddress()
	self.Address = common.Address(address)
	l, _ := source.NextUint32()
	self.Topics = make([]common.Hash, 0, l)
	for i := uint32(0); i < l; i++ {
		h, _ := source.NextHash()
		self.Topics = append(self.Topics, common.Hash(h))
	}
	data, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("StorageLog.Data eof")
	}
	self.Data = data

	return nil
}
type PolyChainListen struct {
	polySdk *chainsdk.PolySDKPro
}

func NewPolyChainListen(urls []string) *PolyChainListen {
	polyListen := &PolyChainListen{}
	sdk := chainsdk.NewPolySDKPro(urls, 1, 0)
	polyListen.polySdk = sdk
	return polyListen
}


type MakeTxParamWithSender struct {
	Sender ontcommon.Address
	ontocccom.MakeTxParam
}

var (
	addrTy, _  = abi.NewType("address", "", nil)
	bytesTy, _ = abi.NewType("bytes", "", nil)
	arguments  = abi.Arguments{
		{Type: addrTy, Name: "Sender"},
		{Type: bytesTy, Name: "Value"},
	}
)

func (this *MakeTxParamWithSender) Serialization() (data []byte, err error) {
	sink := ontcommon.NewZeroCopySink(nil)
	sink.WriteAddress(ontcommon.Address(this.Sender))
	this.MakeTxParam.Serialization(sink)
	data = sink.Bytes()
	return
}

func Compose(tx *msg.Tx) (err error) {
	polyNode:=polynode.New("http://40.115.182.238:40336")
	ontNode:=ontnode.New("http://43.128.242.133:20336")
	v, err := polyNode.GetSideChainMsg(uint64(5555), tx.SrcHeight)
	if err != nil {
		fmt.Println("err GetSideChainMsg",err)
		return  fmt.Errorf("GetSideChainMsg:%s", err)
	}
	if len(v) == 0 {
		msg, err := ontNode.GetCrossChainMsg(uint32(tx.SrcHeight))
		if err != nil {
			fmt.Println("err ontNode.GetCrossChainMsg",err)
			return err
		}
		tx.SrcStateRoot, err = hex.DecodeString(msg)
		if err != nil {
			fmt.Println("err hex.DecodeString(msg)",err)
			return err
		}
	}
	var clientMgr ontSDK1Client.ClientMgr
	clientMgr.NewRpcClient().SetAddress("http://43.128.242.133:20336")
	hashes, err :=clientMgr.GetCrossStatesLeafHashes(float64(tx.SrcHeight))
	if err != nil {
		fmt.Println("err GetCrossStatesLeafHashes",err)
		return  fmt.Errorf("GetCrossStatesLeafHashes:%s", err)
	}
	fmt.Println("hashes.Hashes:",hashes.Hashes)
	eccmAddr := "34d4a23A1FC0C694f0D74DDAf9D8d564cfE2D430"
	eccmAddr =HexStringReverse1(eccmAddr)
	param := ontocccom.MakeTxParam{}
	par,_:=hex.DecodeString(tx.SrcParam)
	err = param.Deserialization(ontcommon.NewZeroCopySource(par))
	if err != nil {
		fmt.Println("err param.Deserialization:", err)
		return err
	}

	ontEccmAddr,err:=ontcommon.AddressFromHexString(eccmAddr)
	makeTxParamWithSender:=&MakeTxParamWithSender{
		ontEccmAddr,
		param,
	}
	itemValue,err:=makeTxParamWithSender.Serialization()
	if err != nil {
		fmt.Println("err makeTxParamWithSender.Serialization:", err)
		return err
	}
	hashesx:=make([]ontcommon.Uint256,0)
	for k,v:=range hashes.Hashes{
		fmt.Println("hashes.Hashes[",k,"] : ",v)
		uint256v,_:=ontcommon.Uint256FromHexString(v)
		hashesx=append(hashesx,uint256v)
	}
	fmt.Println("itemValue:",ontcommon.ToHexString(itemValue))
	fmt.Println("hashesx:",hashesx)
	path, err := merkle.MerkleLeafPath(itemValue, hashesx)
	if err != nil {
		fmt.Println("err  merkle.MerkleLeafPath:", err)
		return err
	}
	proof := &bc.CrossStatesProof{}
	err = json.Unmarshal(path, proof)
	if err != nil {
		fmt.Println("err json.Unmarshal(path, proof)",err)
		return err
	}
	tx.SrcProof, err = hex.DecodeString(proof.AuditPath)
	if err != nil {
		return
	}
	{
		value, _, _, _ := msg.ParseAuditPath(tx.SrcProof)
		if len(value) == 0 {
			return fmt.Errorf("ParseAuditPath got null param")
		}
		param := &ontocccom.MakeTxParam{}
		err = param.Deserialization(ontcommon.NewZeroCopySource(value))
		if err != nil {
			return
		}
		tx.Param = &pplycommon.MakeTxParam{
			TxHash:              param.TxHash,
			CrossChainID:        param.CrossChainID,
			FromContractAddress: param.FromContractAddress,
			ToChainID:           param.ToChainID,
			ToContractAddress:   param.ToContractAddress,
			Method:              param.Method,
			Args:                param.Args,
		}
	}
	return
}


//-tags testnet
//github.com/onto/ontogo-sdk v1.0.0 => /Users/rain/go/gopath/src/github.com/me/fork-onto-go-sdk/ontology-go-sdk
func Test_ontCrossTx(t *testing.T) {
	tx:=scan()
	fmt.Println(tx.SrcHeight)
	err:=Compose(tx)
	fmt.Println("end err:",err)
}
