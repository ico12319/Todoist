package http_helpers

import "encoding/json"

type jsonService struct{}

func NewJsonMarshaller() *jsonService {
	return &jsonService{}
}

func (*jsonService) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
