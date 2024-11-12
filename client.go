package tinyfs_client

import (
	"errors"
	"fmt"
	"github.com/andyzhou/tinyfs_client/define"
	"github.com/andyzhou/tinyfs_client/face"
	"github.com/andyzhou/tinyfs_client/json"
	"sync"
	"sync/atomic"
)

/*
 * client for master service
 */

//global variable
var (
	_client *Client
	_clientOnce sync.Once
)

//face info
type Client struct {
	node *face.Node
	addressArr []string //unique slice
	num int32 //atomic value
	sync.RWMutex
}

//single instance
func GetClient() *Client {
	_clientOnce.Do(func() {
		_client = NewClient()
	})
	return _client
}

//construct
func NewClient() *Client {
	this := &Client{
		node: face.NewNode(),
	}
	return this
}

//quit
func (f *Client) Quit() {
	f.node.Quit()
}

//list file info
func (f *Client) ListFiles(page, pageSize int) (*json.ListFileRespJson, error) {
	//pick rand master node
	node, err := f.node.PickNode()
	if err != nil {
		return nil, err
	}
	if node == nil || node.Client == nil {
		return nil, errors.New("can't get valid node")
	}

	//init list file request
	reqObj := json.NewListFileReqJson()
	reqObj.Page = page
	reqObj.PageSize = pageSize

	//encode request obj
	reqBytes, subErr := reqObj.Encode(reqObj)
	if subErr != nil {
		return nil, subErr
	}

	//gen packet
	pack := node.Client.GenPacket()
	pack.MessageId = define.MessageIdOfListFile
	pack.Data = reqBytes

	//send request to target node
	resp, subErrTwo := node.Client.SendRequest(pack)
	if subErrTwo != nil {
		return nil, subErrTwo
	}
	if resp.ErrCode != define.ErrCodeOfSucceed {
		subErr = fmt.Errorf("read file fialed, code:%v, err:%v\n", resp.ErrCode, resp.ErrMsg)
		return nil, subErr
	}
	//decode origin resp
	respObj := json.NewListFileRespJson()
	respObj.Decode(resp.Data, respObj)
	return respObj, nil
}

//del file info
func (f *Client) DelFiles(shortUrls ...string) error {
	//check
	if shortUrls == nil || len(shortUrls) <= 0 {
		return errors.New("invalid parameter")
	}

	//pick rand master node
	node, err := f.node.PickNode()
	if err != nil {
		return err
	}
	if node == nil || node.Client == nil {
		return errors.New("can't get valid node")
	}

	//init delete file request
	reqObj := json.NewDeleteFileReqJson()
	reqObj.ShortUrls = shortUrls

	//encode request obj
	reqBytes, subErrTwo := reqObj.Encode(reqObj)
	if subErrTwo != nil {
		return subErrTwo
	}

	//gen packet
	pack := node.Client.GenPacket()
	pack.MessageId = define.MessageIdOfDelete
	pack.Data = reqBytes

	//send request to target node
	resp, subErr := node.Client.SendRequest(pack)
	if subErr != nil {
		return subErr
	}
	if resp.ErrCode != define.ErrCodeOfSucceed {
		subErr = fmt.Errorf("read file fialed, code:%v, err:%v\n", resp.ErrCode, resp.ErrMsg)
		return subErr
	}
	return nil
}

//remove file info
func (f *Client) RemoveFiles(shortUrls ...string) error {
	//check
	if shortUrls == nil || len(shortUrls) <= 0 {
		return errors.New("invalid parameter")
	}

	//pick rand master node
	node, err := f.node.PickNode()
	if err != nil {
		return err
	}
	if node == nil || node.Client == nil {
		return errors.New("can't get valid node")
	}

	//init remove file request
	reqObj := json.NewRemoveFileReqJson()
	reqObj.ShortUrls = shortUrls

	//encode request obj
	reqBytes, subErr := reqObj.Encode(reqObj)
	if subErr != nil {
		return subErr
	}

	//gen packet
	pack := node.Client.GenPacket()
	pack.MessageId = define.MessageIdOfRemove
	pack.Data = reqBytes

	//send request to target master node
	resp, subErrTwo := node.Client.SendRequest(pack)
	if subErrTwo != nil {
		return subErrTwo
	}
	if resp.ErrCode != define.ErrCodeOfSucceed {
		subErr = fmt.Errorf("remove file failed, code:%v, err:%v\n", resp.ErrCode, resp.ErrMsg)
		return subErr
	}
	return nil
}

//read file data
func (f *Client) ReadMultiFiles(
		req *json.ReadMultiFilesReqJson,
	) (*json.ReadMultiFilesRespJson, error) {
	//check
	if req == nil || req.ShortUrls == nil || len(req.ShortUrls) <= 0 {
		return nil, errors.New("invalid parameter")
	}
	//pick active master node
	node, err := f.node.PickNode()
	if err != nil {
		return nil, err
	}
	if node == nil || node.Client == nil {
		return nil, errors.New("node client not init")
	}
	reqBytes, _ := req.Encode(req)

	//gen packet
	pack := node.Client.GenPacket()
	pack.MessageId = define.MessageIdOfMultiRead
	pack.Data = reqBytes

	//send request to target node
	resp, subErr := node.Client.SendRequest(pack)
	if subErr != nil {
		return nil, subErr
	}
	if resp.ErrCode != define.ErrCodeOfSucceed {
		subErr = fmt.Errorf("read multi file failed, code:%v, err:%v\n",
			resp.ErrCode, resp.ErrMsg)
		return nil, subErr
	}

	//decode origin resp
	respObj := json.NewReadMultiFilesRespJson()
	respObj.Decode(resp.Data, respObj)
	return respObj, nil
}

func (f *Client) ReadFile(
		req *json.ReadFileReqJson,
	) (*json.ReadFileRespJson, error) {
	//check
	if req == nil || req.ShortUrl == "" {
		return nil, errors.New("invalid parameter")
	}

	//pick active node
	node, err := f.node.PickNode()
	if err != nil {
		return nil, err
	}
	if node == nil || node.Client == nil {
		return nil, errors.New("node client not init")
	}
	reqBytes, _ := req.Encode(req)

	//gen packet
	pack := node.Client.GenPacket()
	pack.MessageId = define.MessageIdOfRead
	pack.Data = reqBytes

	//send request to target node
	resp, subErr := node.Client.SendRequest(pack)
	if subErr != nil {
		return nil, subErr
	}
	if resp.ErrCode != define.ErrCodeOfSucceed {
		subErr = fmt.Errorf("read file failed, code:%v, err:%v\n",
			resp.ErrCode, resp.ErrMsg)
		return nil, subErr
	}

	//decode origin resp
	respObj := json.NewReadFileRespJson()
	respObj.Decode(resp.Data, respObj)
	return respObj, nil
}

//write file data
func (f *Client) WriteFile(
		req *json.WriteFileReqJson,
	) (*json.WriteFileRespJson, error) {
	//check
	if req == nil || req.Name == "" || req.Data == nil {
		return nil, errors.New("invalid parameter")
	}

	//pick active node
	node, err := f.node.PickNode()
	if err != nil {
		return nil, err
	}
	if node == nil || node.Client == nil {
		return nil, errors.New("node client not init")
	}
	reqBytes, _ := req.Encode(req)

	//gen packet
	pack := node.Client.GenPacket()
	pack.MessageId = define.MessageIdOfWrite
	pack.Data = reqBytes

	//send request to target chunk node
	resp, subErr := node.Client.SendRequest(pack)
	if subErr != nil {
		return nil, subErr
	}
	if resp.ErrCode != define.ErrCodeOfSucceed {
		subErr = fmt.Errorf("send file failed, code:%v, err:%v\n", resp.ErrCode, resp.ErrMsg)
		return nil, subErr
	}

	//decode origin response data
	respObj := json.NewWriteFileRespJson()
	respObj.Decode(resp.Data, respObj)
	return respObj, nil
}

//get sub face
func (f *Client) GetNode() *face.Node {
	return f.node
}

//remove master node
//addr format -> host:port
func (f *Client) RemoveNode(addr string) error {
	//check
	if addr == "" {
		return errors.New("invalid parameter")
	}
	nodeObj, _ := f.node.GetNodeByAddr(addr)
	if nodeObj == nil {
		return errors.New("address not exists")
	}
	//remove node
	err := f.node.DelNode(nodeObj.Tag)
	return err
}

//add master node
//addr format -> host:port
func (f *Client) AddNode(addr string, maxMsgSizes ...int) error {
	//check
	if addr == "" {
		return errors.New("invalid parameter")
	}
	if f.checkAddress(addr) {
		return errors.New("address had exists")
	}
	//add new node
	tag := fmt.Sprintf("%v", f.num)
	f.node.AddNode(tag, addr, maxMsgSizes...)
	atomic.AddInt32(&f.num, 1)
	return nil
}

///////////////
//private func
///////////////

//check address
func (f *Client) checkAddress(addr string) bool {
	f.Lock()
	defer f.Unlock()
	for _, v := range f.addressArr {
		if v == addr {
			return true
		}
	}
	return false
}