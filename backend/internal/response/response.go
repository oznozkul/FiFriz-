package response

type Body struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OK(data any) Body {
	return Body{
		Success: true,
		Data:    data,
	}
}

func Fail(code, message string) Body {
	return Body{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	}
}
