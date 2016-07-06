## Ldap2Pop3 is a ldap server ,forward ldap request to pop3Server

* pop3 lib use https://github.com/bytbox/go-pop3
* ldapserver lib use https://github.com/vjeantet/ldapserver



## Configuration:
```
{
    "ListenAddr":"0.0.0.0:10389" ,
    "Pop3ServerTlsAddr": "127.0.0.1:995",   //Pop3 server must be enabled SSL
	"LookupUser":"Root"					//This user will always pass the ldap authentication
}

```
## Install
