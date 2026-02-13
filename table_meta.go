package mt

import (
    "sync/atomic"

    "google.golang.org/protobuf/proto"
)

type TableMeta[T proto.Message] struct {
    atomic.Pointer[T]
}
