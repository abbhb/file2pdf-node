package typeall

type TestS struct {
	Name  string `json:"name"`
	Type  int    `json:"typeall"`
	Fsa   bool   `json:"fsa"`
	Index int    `json:"index"`
}
type PrintDataFileToPDFReq struct {
	Id               *string `json:"id"`
	FileUrl          *string `json:"fileUrl"`
	FilePDFUrl       *string `json:"filePDFUrl"`
	FilePDFUploadUrl *string `json:"filePDFUploadUrl"`
}

type PrintDataFromPDFResp struct {
	Id         *string `json:"id"`
	FilePDFUrl *string `json:"filePDFUrl"`
	PageNums   *int    `json:"pageNums"`
	Status     *int    `json:"status"`
	Message    *string `json:"message"`
}

// 定义一个结构体来传递页数和可能发生的错误
type PageCountResult struct {
	NumPages int
	Err      error
}
