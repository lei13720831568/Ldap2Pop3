// Ldap2Pop3
package main

import (
	"errors"
	//	"fmt"
	"encoding/json"
	pop3 "github.com/bytbox/go-pop3"
	message "github.com/vjeantet/goldap/message"
	ldap "github.com/vjeantet/ldapserver"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"
)

func GetNameFromFilter(fstr string) (string, error) {
	reg := regexp.MustCompile("(?:\\(uid=){1}(.+?)(?:\\)){1}")
	result := reg.FindAllStringSubmatch(fstr, -1)
	if len(result) == 0 {
		return "", errors.New("Not find uid")
	}
	return result[0][1], nil
}

var c Config

func main() {
	//read config
	r, err := os.Open("conf.json")
	if err != nil {
		log.Fatalln(err)
	}
	decoder := json.NewDecoder(r)

	err = decoder.Decode(&c)
	if err != nil {
		log.Fatalln(err)
	}

	//Create a new LDAP Server
	server := ldap.NewServer()
	routes := ldap.NewRouteMux()
	routes.NotFound(handleNotFound)
	routes.Abandon(handleAbandon)
	routes.Bind(handleBind)

	//	routes.Search(handleSearchDSE).
	//		BaseDn("").
	//		Scope(ldap.SearchRequestScopeBaseObject).
	//		Filter("(objectclass=*)").
	//		Label("Search - ROOT DSE")

	//	routes.Search(handleSearchMyCompany).
	//		BaseDn("o=My Company, c=US").
	//		Scope(ldap.SearchRequestScopeBaseObject).
	//		Label("Search - Compagny Root")

	routes.Search(handleSearch).Label("Search - Generic")

	//Attach routes to server
	server.Handle(routes)

	// listen on 10389 and serve
	go server.ListenAndServe(c.ListenAddr)

	// When CTRL+C, SIGINT and SIGTERM signal occurs
	// Then stop server gracefully
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	close(ch)

	server.Stop()

}

func handleNotFound(w ldap.ResponseWriter, r *ldap.Message) {
	switch r.ProtocolOpType() {
	case ldap.ApplicationBindRequest:
		res := ldap.NewBindResponse(ldap.LDAPResultSuccess)
		res.SetDiagnosticMessage("Default binding behavior set to return Success")

		w.Write(res)

	default:
		res := ldap.NewResponse(ldap.LDAPResultUnwillingToPerform)
		res.SetDiagnosticMessage("Operation not implemented by server")
		w.Write(res)
	}
}

func invokePop3Auth(userName string, pwd string) error {
	addr := c.Pop3ServerTlsAddr

	c, err := pop3.DialTLS(addr)
	if err != nil {
		return errors.New("NewClient failed:" + err.Error())
	}

	if err = c.Auth(userName, pwd); err != nil {
		return errors.New("Auth failed" + err.Error())
	}

	c.Noop()

	c.Quit()
	return nil
}

func handleBind(w ldap.ResponseWriter, m *ldap.Message) {
	r := m.GetBindRequest()

	res := ldap.NewBindResponse(ldap.LDAPResultSuccess)
	if r.AuthenticationChoice() == "simple" {
		uName := r.Name()

		log.Println("uName:", uName)
		if string(uName) == c.LookupUser {
			w.Write(res)
			return
		}

		pwd, ok := r.Authentication().(message.OCTETSTRING)

		if !ok {
			res.SetResultCode(ldap.LDAPResultInvalidCredentials)
			res.SetDiagnosticMessage("invalid credentials")
		} else {
			err := invokePop3Auth(string(uName), pwd.String())
			if err != nil {
				res.SetResultCode(ldap.LDAPResultInvalidCredentials)
				res.SetDiagnosticMessage("invalid credentials," + err.Error())
				log.Println("invalid credentials," + err.Error())
			}
		}

		//log.Printf("Bind failed User=%s, Pass=%#v", uName, r.Authentication())

	} else {
		res.SetResultCode(ldap.LDAPResultUnwillingToPerform)
		res.SetDiagnosticMessage("Authentication choice not supported")
	}

	w.Write(res)
}

func handleAbandon(w ldap.ResponseWriter, m *ldap.Message) {
	var req = m.GetAbandonRequest()
	// retreive the request to abandon, and send a abort signal to it
	if requestToAbandon, ok := m.Client.GetMessageByID(int(req)); ok {
		requestToAbandon.Abandon()
		log.Printf("Abandon signal sent to request processor [messageID=%d]", int(req))
	}
}

func handleSearch(w ldap.ResponseWriter, m *ldap.Message) {
	r := m.GetSearchRequest()
	log.Printf("Request BaseDn=%s", r.BaseObject())
	log.Printf("Request Filter=%s", r.Filter())
	log.Printf("Request FilterString=%s", r.FilterString())
	log.Printf("Request Attributes=%s", r.Attributes())
	log.Printf("Request TimeLimit=%d", r.TimeLimit().Int())

	// Handle Stop Signal (server stop / client disconnected / Abandoned request....)
	select {
	case <-m.Done:
		log.Print("Leaving handleSearch...")
		return
	default:
	}

	uName, err := GetNameFromFilter(r.FilterString())
	if err == nil {
		log.Println("search uName done:", uName)
		e := ldap.NewSearchResultEntry(uName)
		e.AddAttribute("mail", "")
		e.AddAttribute("cn", message.AttributeValue(uName))
		w.Write(e)
	}

	res := ldap.NewSearchResultDoneResponse(ldap.LDAPResultSuccess)
	w.Write(res)

}
