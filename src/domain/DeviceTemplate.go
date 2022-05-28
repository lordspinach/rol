package domain

//DeviceTemplate represents yaml device template as a structure
type DeviceTemplate struct {
	//Name template name
	Name string
	//Model device model
	Model string
	//Manufacturer device manufacturer
	Manufacturer string
	//Description template description
	Description string
	//CPUCount count of cpus
	CPUCount int
	//CPUModel model of cpu
	CPUModel string
	//RAM the amount of RAM in GB
	RAM int
	//NetworkInterfaces slice of device network interfaces
	NetworkInterfaces []DeviceTemplateNetworkInterface
	//Control describes how we control the device
	Control DeviceTemplateControlDesc
	//DiscBootStages slice of boot stage templates for disk boot
	DiscBootStages []BootStageTemplate
	//NetBootStages slice of boot stage templates for net boot
	NetBootStages []BootStageTemplate
	//USBBootStages slice of boot stage templates for usb boot
	USBBootStages []BootStageTemplate
}

//DeviceTemplateNetworkInterface is a structure that stores information about network interface
type DeviceTemplateNetworkInterface struct {
	//Name of network interface. This field is unique within device template network interfaces
	Name string
	//NetBoot flags whether the interface can be loaded over the network
	NetBoot bool
	//POEIn only one network interface can be mark as POEIn
	POEIn bool
	//Management only one network interface can be mark as management
	Management bool
}

//DeviceTemplateControlDesc is a structure that stores information
//about how to control the device in different situations
type DeviceTemplateControlDesc struct {
	//Emergency how to control device power in case of emergency. As example: POE(For Rpi4), IPMI, ILO or PowerSwitch
	Emergency string
	//Power how to control device power. As example: POE(For Rpi4), IPMI, ILO or PowerSwitch
	Power string
	//NextBoot how to change next boot device. As example: IPMI, ILO or NONE.
	//For example, NONE is used for Rpi4, we control next boot by u-boot files in boot stages.
	NextBoot string
}

//BootStageTemplateFile is a structure that stores the path to the bootstrap file
type BootStageTemplateFile struct {
	//ExistingFileName file name is a real full file path with name on the disk.
	//This path is relative from app directory
	ExistingFileName string
	//VirtualFileName virtual file name is relative from /<mac-address>/
	VirtualFileName string
}

//BootStageTemplate boot stage can be overwritten in runtime by device entity or by device rent entity.
//BootStageTemplate converts to BootStage for device, then we create device entity.
type BootStageTemplate struct {
	//Name of boot stage template
	Name string
	//Description of boot stage template
	Description string
	//Action for this boot stage.
	//Can be: File, CheckPowerSwitch, EmergencyPowerOff,
	//PowerOff, EmergencyPowerOn, PowerOn,
	//CheckManagement
	//
	//For File action:
	//	A stage can only be marked complete if all files have
	//	been downloaded by the device via TFTP or DHCP,
	//	after which the next step can be loaded.
	Action string
	//Files slice of files for boot stage
	Files []BootStageTemplateFile
}
