package face

import (
	"errors"
	"github.com/andyzhou/tinyrpc"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * general node face
 */
const (
	DefaultMaxMsgSize = 1024 * 1024 * 10 //10MB
	DefaultNodeConnDelaySeconds = 5 //xx seconds
	DefaultNodeCheckRate = 5 //xx seconds
)

//one node info
type OneNode struct {
	Tag string
	Address string
	Client *tinyrpc.Client
	Connected bool
}

//face info
type Node struct {
	nodeMap sync.Map //tag -> *OneNode
	tags []string
	nodes int32
	ticker *time.Ticker
	closeChan chan bool
	sync.RWMutex
}

//construct
func NewNode() *Node {
	this := &Node{
		nodeMap: sync.Map{},
		tags: []string{},
		closeChan: make(chan bool, 1),
	}
	this.interInit()
	return this
}

//quit
func (f *Node) Quit() {
	if f.closeChan != nil {
		close(f.closeChan)
	}
}

//get tags
func (f *Node) GetTags() []string {
	return f.tags
}

//del one node
func (f *Node) DelNode(tag string) error {
	//check
	if tag == "" {
		return errors.New("invalid parameter")
	}

	//get one
	node, err := f.GetNode(tag)
	if err != nil {
		return err
	}
	if node == nil {
		return nil
	}
	if node.Client != nil {
		node.Client.Quit()
	}

	//remove from map
	f.nodeMap.Delete(tag)
	atomic.AddInt32(&f.nodes, -1)
	if f.nodes < 0 {
		atomic.StoreInt32(&f.nodes, 0)
	}

	//remove tag from slice
	hitIdx := -1
	for idx, val := range f.tags {
		if val == tag {
			hitIdx = idx
			break
		}
	}
	if hitIdx >= 0 {
		//remove element
		f.Lock()
		defer f.Unlock()
		f.tags = append(f.tags[:hitIdx], f.tags[hitIdx+1:]...)
	}
	return nil
}

//get all nodes
func (f *Node) GetAllNode() map[string]*OneNode {
	result := make(map[string]*OneNode)
	sf := func(k, v interface{}) bool {
		node, ok := v.(*OneNode)
		if ok && node != nil {
			result[node.Tag] = node
		}
		return true
	}
	f.nodeMap.Range(sf)
	return result
}

//get node by tag
func (f *Node) GetNode(tag string) (*OneNode, error) {
	//check
	if tag == "" {
		return nil, errors.New("invalid parameter")
	}
	v, ok := f.nodeMap.Load(tag)
	if !ok || v == nil {
		return nil, nil
	}
	node, subOk := v.(*OneNode)
	if !subOk || node == nil {
		return nil, errors.New("invalid node data")
	}
	return node, nil
}

//get node by address
func (f *Node) GetNodeByAddr(addr string) (*OneNode, error) {
	var (
		target *OneNode
	)
	//check
	if addr == "" {
		return nil, errors.New("invalid parameter")
	}
	//loop check
	sf := func(k, v interface{}) bool {
		tag, _ := k.(string)
		obj, _ := v.(*OneNode)
		if tag != "" && obj != nil {
			if obj.Address == addr {
				//found it
				target = obj
				return false
			}
		}
		return true
	}
	f.nodeMap.Range(sf)
	return target, nil
}

//pick rand node
func (f *Node) PickNode() (*OneNode, error) {
	if f.nodes <= 0 {
		return nil, errors.New("no any node")
	}
	rand.Seed(time.Now().UnixNano())
	randIdx := rand.Intn(int(f.nodes))
	tag := f.tags[randIdx]
	node, err := f.GetNode(tag)
	return node, err
}

//add node
func (f *Node) AddNode(tag, address string, maxMsgSizes ...int) error {
	var (
		maxMsgSize int
	)
	//check
	if tag == "" || address == "" {
		return errors.New("invalid parameter")
	}
	v, _ := f.GetNode(tag)
	if v != nil {
		return nil
	}

	//init new node
	newNode := &OneNode{
		Tag: tag,
		Address: address,
	}

	//save node info
	defer func() {
		//sync into map
		f.nodeMap.Store(tag, newNode)
		atomic.AddInt32(&f.nodes, 1)

		//update tag slice
		f.Lock()
		defer f.Unlock()
		f.tags = append(f.tags, tag)
	}()

	//detect
	if maxMsgSizes != nil && len(maxMsgSizes) > 0 {
		maxMsgSize = maxMsgSizes[0]
	}
	if maxMsgSize <= 0 {
		maxMsgSize = DefaultMaxMsgSize
	}

	//get client para
	clientPara := &tinyrpc.ClientPara{
		MaxMsgSize: maxMsgSize,
	}

	//init new client
	client := tinyrpc.NewClient(clientPara)
	client.SetAddress(address)
	client.SetServerNodeDownCallBack(f.cbForServerNodeDown)

	//set client obj
	newNode.Client = client

	//connect server
	err := client.ConnectServer()
	if err != nil {
		return err
	}

	//set new node
	newNode.Connected = true
	return nil
}

////////////////
//private func
////////////////

//cb for server node down
func (f *Node) cbForServerNodeDown(serverAddr string) error {
	//check
	if serverAddr == "" {
		return errors.New("invalid parameter")
	}

	//try re-connect server in son process
	sf := func() {
		rpcNode, _ := f.getNodeByAddr(serverAddr)
		if rpcNode != nil {
			//force close rpc client
			if rpcNode.Client != nil {
				rpcNode.Client.Quit()
			}
			//re-connect target server force
			//run in son process
			go f.reConnectDownedServerNode(rpcNode)
		}
	}
	sf()
	return nil
}

//re-connect downed server node
func (f *Node) reConnectDownedServerNode(serverNode *OneNode) error {
	var (
		finalClient *tinyrpc.Client
	)
	//check
	if serverNode == nil || serverNode.Address == "" {
		return errors.New("invalid parameter")
	}

	//get key data
	nodeAddr := serverNode.Address

	//loop connect server
	for {
		//init new rpc client
		newClient := tinyrpc.NewClient()
		err := newClient.SetAddress(nodeAddr)
		newClient.SetServerNodeDownCallBack(f.cbForServerNodeDown)

		//connect server
		err = newClient.ConnectServer()
		if err != nil {
			log.Printf("connect rpc server %v failed, err:%v\n", nodeAddr, err.Error())
			newClient.Quit()
			time.Sleep(time.Second * DefaultNodeConnDelaySeconds)
		}else{
			//connect success
			log.Printf("connect rpc server %v success..\n", nodeAddr)
			finalClient = newClient
			break
		}
	}

	//update active client
	serverNode.Client = finalClient
	f.nodeMap.Store(serverNode.Tag, serverNode)
	return nil
}

//get node by address
func (f *Node) getNodeByAddr(addr string) (*OneNode, error) {
	var (
		target *OneNode
	)
	//check
	if addr == "" {
		return nil, errors.New("invalid parameter")
	}
	//loop check
	sf := func(k, v interface{}) bool {
		tag, _ := k.(string)
		obj, _ := v.(*OneNode)
		if tag != "" && obj != nil {
			if obj.Address == addr {
				//found it
				target = obj
				return false
			}
		}
		return true
	}
	f.nodeMap.Range(sf)
	return target, nil
}

//check inter node
func (f *Node) checkNodes() {
	//check
	if f.nodes <= 0 {
		return
	}
	//loop check
	//if client not connected, just re-connect
	sf := func(k, v interface{}) bool {
		nodeObj, ok := v.(*OneNode)
		if ok && nodeObj != nil && !nodeObj.Connected {
			err := nodeObj.Client.ConnectServer()
			if err == nil {
				nodeObj.Connected = true
			}
		}
		return true
	}
	f.nodeMap.Range(sf)
}

//inter ticker checker
func (f *Node) nodeCheckTicker() {
	var (
		m any = nil
	)
	defer func() {
		if err := recover(); err != m {
			panic(any(err))
		}
	}()

	for {
		select {
		case <- f.ticker.C:
			{
				f.checkNodes()
			}
		case <- f.closeChan:
			return
		}
	}
}

//inter init
func (f *Node) interInit() {
	//init ticker
	f.ticker = time.NewTicker(time.Duration(DefaultNodeCheckRate) * time.Second)
	go f.nodeCheckTicker()
}