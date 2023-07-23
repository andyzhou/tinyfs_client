package face


import (
	"errors"
	"github.com/andyzhou/tinyrpc"
	"math/rand"
	"sync"
	"sync/atomic"
)

/*
 * general node face
 */

//one node info
type OneNode struct {
	Tag string
	Address string
	Client *tinyrpc.Client
}

//face info
type Node struct {
	nodeMap sync.Map
	tags []string
	nodes int32
	sync.RWMutex
}

//construct
func NewNode() *Node {
	this := &Node{
		nodeMap: sync.Map{},
		tags: []string{},
	}
	return this
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

//pick rand node
func (f *Node) PickNode() (*OneNode, error) {
	if f.nodes <= 0 {
		return nil, errors.New("no any node")
	}
	randIdx := rand.Intn(int(f.nodes))
	tag := f.tags[randIdx]
	node, err := f.GetNode(tag)
	return node, err
}

//add node
func (f *Node) AddNode(tag, address string) error {
	//check
	if tag == "" || address == "" {
		return errors.New("invalid parameter")
	}
	v, _ := f.GetNode(tag)
	if v != nil {
		return nil
	}
	//init new client
	client := tinyrpc.NewClient()
	client.SetAddress(address)
	err := client.ConnectServer()
	if err != nil {
		return err
	}
	//init new node
	newNode := &OneNode{
		Tag: tag,
		Address: address,
		Client: client,
	}

	//sync into map
	f.nodeMap.Store(tag, newNode)
	atomic.AddInt32(&f.nodes, 1)

	//update tag slice
	f.Lock()
	defer f.Unlock()
	f.tags = append(f.tags, tag)
	return nil
}