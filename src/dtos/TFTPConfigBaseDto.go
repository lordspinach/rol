package dtos

//TFTPConfigBaseDto TFTP config base dto
type TFTPConfigBaseDto struct {
	//Address TFTP server IP address
	Address string
	//Port TFTP server port
	Port string
	//Enabled TFTP server startup status
	Enabled bool
}
