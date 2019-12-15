package protocols

// ErrorMessage uses for describe the error message and give it to the front end
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
