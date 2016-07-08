## Ldap2Pop3 is a ldap server ,forward ldap request to pop3Server

* pop3 lib use https://github.com/bytbox/go-pop3
* ldapserver lib use https://github.com/vjeantet/ldapserver

---------------
## Install

* install golang
* config GOPATH

```
go get github.com/lei13720831568/Ldap2Pop3
go install github.com/lei13720831568/Ldap2Pop3
cp $GOPATH/src/github.com/lei13720831568/Ldap2Pop3/conf.json  $GOPATH/bin/conf.json

```

### Configuration:
```
{
  "ListenAddr":"0.0.0.0:10389" ,
  "Pop3ServerTlsAddr": "127.0.0.1:995",   //Pop3 server must be enabled SSL
  "LookupUser":"Root"					//This user will always pass the ldap authentication
}

```
------------------
## Run on docker
Minimal docker image only 7.393 MB

* install golang
* config GOPATH
* config conf.json

```
go get github.com/lei13720831568/Ldap2Pop3
cd $GOPATH/src/github.com/lei13720831568/Ldap2Pop3
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .
cp Ldap2Pop3 ./docker/Ldap2Pop3
cp conf.json ./docker/conf.json
cd ./docker
docker build -t ldap2pop3 ./docker/
docker run -d -p 10389:10389 --name ldap2pop3 ldap2pop3
```
