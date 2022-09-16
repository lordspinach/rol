package infrastructure

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pin/tftp/v3"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
	"rol/app/errors"
	"rol/app/utils"
	"rol/domain"
)

type tftpServer struct {
	server    *tftp.Server
	config    domain.TFTPConfig
	paths     *[]domain.TFTPPathRatio
	callbacks []func(addr net.UDPAddr, virtualPath, actualPath string) error
	logger    *logrus.Logger
	enabled   bool
}

//TFTPServerManager tftp server manager struct
type TFTPServerManager struct {
	servers []*tftpServer
	logger  *logrus.Logger
}

//NewTFTPServerManager constructor for TFTPServerManager
func NewTFTPServerManager(log *logrus.Logger) *TFTPServerManager {
	return &TFTPServerManager{
		servers: *new([]*tftpServer),
		logger:  log,
	}
}

//CreateTFTPServer create new TFTP server on host
//
//Params:
//	config - tftp config
func (t *TFTPServerManager) CreateTFTPServer(config domain.TFTPConfig) {
	server := tftpServer{
		server:    nil,
		config:    config,
		paths:     &[]domain.TFTPPathRatio{},
		callbacks: []func(addr net.UDPAddr, virtualPath, actualPath string) error{},
		logger:    t.logger,
	}
	server.server = tftp.NewServer(
		func(filename string, rf io.ReaderFrom) error {
			actualPath := ""
			for _, ratio := range *server.paths {
				if ratio.VirtualPath == filename {
					actualPath = ratio.ActualPath
				}
			}
			if actualPath == "" {
				return errors.NotFound.New("no such file or directory")
			}
			file, err := os.Open(actualPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return err
			}
			_, err = rf.ReadFrom(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return err
			}
			return nil
		}, nil)
	t.servers = append(t.servers, &server)
}

//StartTFTPServer start TFTP server on host
//
//Params:
//	id - tftp config ID
func (t *TFTPServerManager) StartTFTPServer(id uuid.UUID) {
	for _, server := range t.servers {
		if server.config.ID == id {
			go func() {
				server.enabled = true
				err := server.server.ListenAndServe(fmt.Sprintf("%s:%s", server.config.Address, server.config.Port))
				if err != nil {
					server.enabled = false
				}
			}()
		}
	}
}

//StopTFTPServer stop TFTP server on host
//
//Params:
//	id - tftp config ID
func (t *TFTPServerManager) StopTFTPServer(id uuid.UUID) {
	for _, server := range t.servers {
		if server.config.ID == id {
			if server.enabled {
				server.server.Shutdown()
				server.enabled = false
			}
		}
	}
}

//UpdatePaths update paths ratio on TFTP server
//
//Params:
//	id - tftp config ID
//	paths - paths ratio
func (t *TFTPServerManager) UpdatePaths(id uuid.UUID, paths []domain.TFTPPathRatio) {
	for _, server := range t.servers {
		if server.config.ID == id {
			server.paths = &paths
		}
	}
}

//ServerIsRunning get server startup status
//
//Params:
//	id - tftp config ID
//Return:
//	bool - true if server is running, otherwise false
func (t *TFTPServerManager) ServerIsRunning(id uuid.UUID) bool {
	for _, server := range t.servers {
		if server.config.ID == id {
			return server.enabled
		}
	}
	return false
}

//DeleteServer removes server object from servers slice
//
//Params:
//	id - server ID
func (t *TFTPServerManager) DeleteServer(id uuid.UUID) {
	for index, s := range t.servers {
		if s.config.ID == id {
			t.servers = utils.RemoveFromSlice(t.servers, index)
		}
	}
}
