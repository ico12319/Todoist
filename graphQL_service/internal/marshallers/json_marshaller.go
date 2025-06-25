package marshallers

import "encoding/json"

type jsonMarshaller struct{}

func NewJsonMarshaller() *jsonMarshaller {
	return &jsonMarshaller{}
}

func (*jsonMarshaller) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
