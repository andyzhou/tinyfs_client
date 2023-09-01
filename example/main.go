package main

import (
	"github.com/andyzhou/tinyfs_client"
	"github.com/andyzhou/tinyfs_client/json"
	"io/ioutil"
	"log"
	"sync"
)

/*
 * client example code
 */

var (
	nodeAddr = "localhost:7100"
	localFile = "./test.txt"
	fileShortUrl = "sIxFt4"
)

//delete file data request
func deleteReq(c *tinyfs_client.Client) {
	//send delete request
	err := c.DelFiles(fileShortUrl)
	log.Println("err:", err)
}

//remove file info request
func removeReq(c *tinyfs_client.Client) {
	//send remove request
	err := c.RemoveFiles(fileShortUrl)
	log.Println("err:", err)
}

//list file info request
func listFileReq(c *tinyfs_client.Client) {
	//send list file request
	page := 1
	pageSize := 10
	fileList, err := c.ListFiles(page, pageSize)
	log.Printf("list file, err:%v\n", err)
	if fileList != nil && fileList.List != nil {
		for _, v := range fileList.List {
			log.Printf("list file, shortUrl:%v, name:%v, size:%v\n",
				v.ShortUrl, v.Name, v.Size)
		}
	}
}


//read file request
func readReq(c *tinyfs_client.Client) {
	//read file request
	req := json.NewReadFileReqJson()
	req.ShortUrl = fileShortUrl

	//send read request
	resp, err := c.ReadFile(req)
	if err != nil {
		log.Println("err:", err)
		return
	}
	log.Println("file name:", resp.Name, ", size:", resp.Size, ", data:", string(resp.Data))
}

//write file request
func writeReq(c *tinyfs_client.Client)  {
	//read local file
	fileByte, err := ioutil.ReadFile(localFile)
	if err != nil {
		log.Println(err)
		return
	}

	//write file request
	req := json.NewWriteFileReqJson()
	req.Name = localFile
	req.Size = int64(len(fileByte))
	req.Data = fileByte

	//send write request
	resp, err := c.WriteFile(req)
	log.Println("resp short url:", resp.ShortUrl, ", err:", err)
}

func main() {
	var (
		wg sync.WaitGroup
	)
	//init client
	client := tinyfs_client.NewClient()

	//add  node
	err := client.AddNode(nodeAddr)
	if err != nil {
		log.Println(err)
		return
	}

	wg.Add(1)

	//send request
	//writeReq(client)
	//readReq(client)
	//delReq(client)
	listFileReq(client)

	wg.Wait()
}
