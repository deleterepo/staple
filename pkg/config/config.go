package config

import "errors"

// Config contains configuration options for the inner server of Staple.
type Config struct {
	AutoTLS        bool
	CacheDir       string
	ServerKeyPath  string
	ServerCrtPath  string
	Port           string
	Hostname       string
	GlobalTokenKey string
	Database       struct {
		ConnectionURL string
	}
	Mailer struct {
		Domain string
		ApiKey string
	}
}

// Opts represents server side configuration for Staple.
var Opts = Config{}

// Message represents an error message.
type Message struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// ApiError wraps a message and a code into a struct for JSON parsing.
func ApiError(m string, code int, err error) Message {
	if err == nil {
		err = errors.New("")
	}
	return Message{
		Code:    code,
		Message: m,
		Error:   err.Error(),
	}
}
