# RoL is Rack of Labs

RoL is an open source platform for managing bare metal servers in your lab.
We use a REST API with a CRUD model to manage configuration and interact with the system.
RoL is currently under **active development**, you can follow the current status here.

## Current status to MVP

- [x] Multi-layer architecture
- [x] Logging to database
- [x] Logs reading from API
- [x] Custom typed errors
- [x] Ethernet Switches configuration management
- [x] Ethernet Switches VLAN's and POE management (only a few switch models)
- [x] Host VLAN's and bridges management
- [x] Host network configuration saver and recover
- [x] Device templates
- [x] DHCP servers management
- [ ] TFTP servers management
- [ ] Devices management
- [ ] Projects management
- [ ] iPXE provisioning

## Install Dependencies

The following steps are required to build and run EVE from source:

### Get Go 1.18.x or newest

`https://golang.org/dl/`

### Install MySQL database

You can install MySQL database locally to you distro or use docker.
We need a MySQL user with the ability to log in with a password and with the rights to create databases and create tables in them.

## How to build

`cd src && go mod tidy && go build`

## How to run

RoL by default read config with name appConfig.yml from pwd directory.

1. You need setup correct login and passwords to you MySQL database in appConfig.yml in `database:entity` and `database:log` sections.
2. Add rights for network management.
   1. Ubuntu: `sudo setcap cap_net_admin+ep ./rol`
   2. Others: you can run `./rol` as root.
3. Run RoL binary.
`./rol`
4. If all ok the last output string will be: `[GIN-debug] Listening and serving HTTP on localhost:8080`
5. Go to the [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) to read API swagger documentation.

## For developers

A typical multi-layer architecture is implemented.

### Folders structure

    .
    ├── docs                # Docs files
    │   ├── plantuml        # Struct diagrams in puml and svg formats
    ├── src                 # Source code
    │   ├── tests           # Unit tests
    │   ├── domain          # Entities
    │   ├── dtos            # DTO's is Data transfer objects
    │   ├── app             # Application logic
    │   │   ├── errors      # Custom errors implementation
    │   │   ├── interfaces  # All defined interfaces
    │   │   ├── mappers     # DTO to Entity converters
    │   │   ├── services    # Entities management logic
    │   │   ├── utils       # Utils and simple helpers
    │   │   ├── validators  # DTO's validators
    │   ├── webapi          # HTTP WEP API application
    │   │   ├── controllers # API controllers
    │   │   ├── swagger     # Swagger auto-generated docs
    │   ├── infrastructure  # Implemenatations
    └── ...
