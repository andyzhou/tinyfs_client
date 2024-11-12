package json

/*
 * file info json
 */

//file info for master side
//saved in global storage for query, db or redis?
type FileInfo struct {
	ShortUrl  string `json:"shortUrl"` //unique key
	Name      string `json:"name"`
	Type      string `json:"type"`
	Size      int64  `json:"size"`
	Md5       string `json:"md5"`
	ChunkNode string `json:"chunkNode"` //chunk node tag
	CreateAt  int64  `json:"createAt"`
	BaseJson
}

//construct
func NewFileInfo() *FileInfo {
	this := &FileInfo{}
	return this
}