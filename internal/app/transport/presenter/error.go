package presenter

import "template-go-with-silverinha-file-genarator/internal/app/transport/outbound"

func Error(message string, params interface{}) *outbound.Error {
	if params == nil {
		params = make(map[string]interface{}, 0)
	}

	return &outbound.Error{
		Status:  "error",
		Message: message,
		Params:  params,
	}
}
