package mt

import (
    "strings"

    "github.com/spf13/cast"
)

func TableCacheKey(m ITableManager, args ...any) string {
    var sb strings.Builder
    sb.WriteString(m.TableName())
    for _, v := range args {
        sb.WriteString(":")
        sb.WriteString(cast.ToString(v))
    }
    return sb.String()
}
