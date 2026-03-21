package api

func singleFieldError(fieldName string, message string) map[string]string {
	if fieldName == "" || message == "" {
		return nil
	}

	return map[string]string{
		fieldName: message,
	}
}

func credentialFieldErrors(identityField string, message string) map[string]string {
	fieldErrors := map[string]string{
		"password": message,
	}
	if identityField != "" {
		fieldErrors[identityField] = message
	}

	return fieldErrors
}
