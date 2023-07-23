package json

/*
 * file request json
 */

//del file
type DelFileReqJson struct {
	ShortUrl string `json:"shortUrl"`
	BaseJson
}

//read file
type ReadFileReqJson struct {
	ShortUrl string `json:"shortUrl"`
	Start int64 `json:"start"`
	Size int64 `json:"size"`
	BaseJson
}

type ReadFileRespJson struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int64 `json:"size"`
	Data []byte `json:"data"`
	BaseJson
}

//write file
type WriteFileReqJson struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int64 `json:"size"`
	Data []byte `json:"data"`
	BaseJson
}

type WriteFileRespJson struct {
	ShortUrl string `json:"shortUrl"`
	BaseJson
}

//construct
func NewDelFileReqJson() *DelFileReqJson {
	this := &DelFileReqJson{}
	return this
}

func NewWriteFileReqJson() *WriteFileReqJson {
	this := &WriteFileReqJson{}
	return this
}
func NewWriteFileRespJson() *WriteFileRespJson {
	this := &WriteFileRespJson{}
	return this
}

func NewReadFileReqJson() *ReadFileReqJson {
	this := &ReadFileReqJson{}
	return this
}
func NewReadFileRespJson() *ReadFileRespJson {
	this := &ReadFileRespJson{
		Data: []byte{},
	}
	return this
}