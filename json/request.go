package json

/*
 * file request json
 */

//del file
type DelFileReqJson struct {
	ShortUrls []string `json:"shortUrls"`
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

//read multi files
type ReadMultiFilesReqJson struct {
	ShortUrls []string `json:"shortUrls"`
	BaseJson
}

type ReadMultiFilesRespJson struct {
	Files map[string]*ReadFileRespJson `json:"files"`
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
	this := &DelFileReqJson{
		ShortUrls: []string{},
	}
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

func NewReadMultiFilesReqJson() *ReadMultiFilesReqJson {
	this := &ReadMultiFilesReqJson{
		ShortUrls: []string{},
	}
	return this
}
func NewReadMultiFilesRespJson() *ReadMultiFilesRespJson {
	this := &ReadMultiFilesRespJson{
		Files: map[string]*ReadFileRespJson{},
	}
	return this
}