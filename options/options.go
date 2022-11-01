package options

type RelayOptions struct {
	RelayHost string
	RelayPort string
}

type SenderOptions struct {
	ShareCode     string
	Relay         string
	HashAlgorithm string
	Zip           bool
	LocalNetwork  bool
}

type ReceiverOptions struct {
	ShareCode    string
	Relay        string
	OutPath      string
	Zip          bool
	LocalNetwork bool
}

type GrpcServerOptions struct {
	Address string
	Ports   string
}
