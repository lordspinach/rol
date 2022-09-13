package domain

//TFTPConfig TFTP config entity
type TFTPConfig struct {
	Entity
	//Address TFTP server IP address
	Address string
	//Port TFTP server port
	Port string
	//Enabled TFTP server startup status
	Enabled bool
}
