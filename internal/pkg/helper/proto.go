package helper

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"log"
	"time"
)

func ConvertProtoTimestampOrNow(t *timestamp.Timestamp) time.Time {
	if t == nil || !t.IsValid() {
		log.Printf("invalid timestamppb: %v", t.CheckValid())
		return time.Now()
	}

	return t.AsTime()
}
