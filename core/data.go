package core

type DataType int32

const (
	DataTypeOperation DataType = 0
	DataTypeMessage   DataType = 1
)

type Operation int32

const (
	OperationLogin     Operation = 0
	OperationLogout    Operation = 1
	OperationKeepAlive Operation = 2
)

type Data struct {
	FromId    int32 // 0: Server
	ToId      int32 // -1: ALL Users
	Time      string
	DataType  DataType
	Message   string
	Operation Operation
}
