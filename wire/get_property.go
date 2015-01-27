package wire

type GetPropertyRequest struct {
	Handle *string //`json:"handle,omitempty"`
	Key    *string //`json:"key,omitempty"`
}

func NewGetPropertyRequest(handle string, name string) *GetPropertyRequest {
	return &GetPropertyRequest{
		Handle: PString(handle),
		Key:    PString(name),
	}

}

type GetPropertyResponse struct {
	Value *string // `json:"value,omitempty"`
}
