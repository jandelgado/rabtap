package rabtap

// ExternalAuth satisfies amqp.Authentication interface and is used
// with mtls
type ExternalAuth struct {
}

func (auth *ExternalAuth) Mechanism() string {
	return "EXTERNAL"
}

func (auth *ExternalAuth) Response() string {
	return "\000*\000*"
}
