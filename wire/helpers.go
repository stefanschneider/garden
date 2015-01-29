package wire

// pString copies a string and returns the address of the copy.
func pString(src string) *string {
	return &src
}

// pBool copies a bool and returns the address of the copy.
func pBool(b bool) *bool {
	return &b
}

func optString(src string) *string {
	if src == "" {
		return nil
	}
	return pString(src)
}

func stringOpt(src *string) string {
	if src == nil {
		return ""
	}
	return *src
}
