package mt

import (
    "fmt"
    "strings"
)

func TableCacheKey(m ITableManager, args ...any) string {
    var s []string
    if len(args) > 0 {
        for _, v := range args {
            s = append(s, fmt.Sprintf("%v", v))
        }
    }
    return fmt.Sprintf("%s:%s", m.TableName(), strings.Join(s, ":"))
}
