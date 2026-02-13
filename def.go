package mt

const (
    TableSizeDefault = 500
    TableKeyDefault  = "CTables"
)

type ITableManager interface {
    TableName() string
    TableCrc() uint32
    Load(tableCrc uint32, tableData []byte) error
    Watch(tableCrc uint32, tableData []byte) error
}
