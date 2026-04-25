package resp

type MessageType int

const (
	SimpleStringMessageType MessageType = iota
	SimpleErrorMessageType
	IntegerMessageType
	BulkStringMessageType
	ArrayMessageType
)

func (i MessageType) String() string {
	return [...]string{"Simple String", "Simple Error", "Integer", "Bulk String", "Array"}[i]
}

func (i MessageType) Index() int {
	return int(i)
}
