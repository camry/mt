package mt

type ITableManager interface {
    TableName() string
    TableCrC() uint32
    Load(tableCrC uint32, tableData []byte) error
    Watch(tableCrC uint32, tableData []byte) error
}
