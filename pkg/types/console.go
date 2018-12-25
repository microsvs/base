package types

type ConsoleInfo struct {
	Mobile string `mapstructure:"mobile" msgpack:"mobile" json:"mobile"`
	UserID string `mapstructure:"userid" msgpack:"userid" json:"userid"`
	Client string `mapstructure:"client" msgpack:"client" json:"client"`
}

// custom error response format return http request
type CustomError struct {
	ErrCode int    `json:"code"`
	ErrMsg  string `json:"message"`
}
