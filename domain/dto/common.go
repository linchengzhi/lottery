package dto

import "encoding/json"

type CommonReq struct {
	RequestId   string          `json:"requestId"`
	RequestTime int64           `json:"requestTime"`
	Data        json.RawMessage `json:"data"`
}
