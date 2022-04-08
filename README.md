## Hero Emulator
### Introduction
Hero Emulator is an open source project which has been created for educational purposes, does not purpose making profit and contain any copyrighted content by any corporations. It has been designed to be executed in a kubernetes cluster and behaves as a server emulator.

### Requirements
* Go >= 1.11
* PostgreSQL
* DBeaver [Recommended] or pgAdmin
* Git Bash (for Windows)
* Redis [Optional]
* K8s cluster [Optional]
* Docker repository [Optional]

### Environment
The following environment variables have to be set on the running environment.

* POSTGRES_HOST
* POSTGRES_PORT
* POSTGRES_USER
* POSTGRES_PASSWORD
* POSTGRES_DB
* SERVER_IP
* DROP_RATE
* EXP_RATE
* PROXY_ENABLED [Optional]
* REDIS_HOST [Optional]
* REDIS_PORT [Optional]
* REDIS_PASSWORD [Optional]
* REDIS_SCHEME [Optional]

### Installation
Source code can be compiled by `go build` command, and the output can be used to start serving directly. However, using the executable binary itself may end up with undesired results. Instead, deploying into a kubernetes cluster is strongly recommended.

ALL CREDITS TO THE ORIGINAL EMULATOR DRAGON LEGEND!
