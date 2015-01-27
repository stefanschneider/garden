package wire

// PString copies a string and returns the address of the copy.
func PString(src string) *string {
	return &src
}

func OptString(src string) *string {
	if src == "" {
		return nil
	}
	return PString(src)
}

// PBool copies a bool and returns the address of the copy.
func PBool(b bool) *bool {
	return &b
}
