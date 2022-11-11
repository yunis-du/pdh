package options

type RelayOptions struct {
	RelayHost string
	RelayPort string
}

type SenderOptions struct {
	ShareCode    string
	Relay        string
	Zip          bool
	LocalNetwork bool
	LocalPort    string
}

type ReceiverOptions struct {
	ShareCode    string
	Relay        string
	OutPath      string
	Zip          bool
	LocalNetwork bool
	LocalPort    string
}

type GrpcServerOptions struct {
	Address string
	Ports   string
}
