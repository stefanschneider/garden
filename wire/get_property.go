package wire

type GetPropertyRequest struct {
	Handle *string //`json:"handle,omitempty"`
	Key    *string //`json:"key,omitempty"`
}

func NewGetPropertyRequest(handle string, name string) *GetPropertyRequest {
	return &GetPropertyRequest{
		Handle: pString(handle),
		Key:    pString(name),
	}

}

type GetPropertyResponse struct {
	Value *string // `json:"value,omitempty"`
}

func NewGetPropertyResponse(value string) *GetPropertyResponse {
	return &GetPropertyResponse{Value: &value}
}
