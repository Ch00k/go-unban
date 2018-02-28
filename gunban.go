package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/Ch00k/gunban/f2bclient"
)

const tmpl = `<html><body>
<p>{{ .IP }} {{if not .Banned}}not{{end}} banned</p>
{{if .Banned}}
	<form action="/" method=post>
		<input type=hidden name=ip value="{{ .IP }}">
		<input type=submit value=unban>
	</form>
{{end}}
</body></html>`

func unban(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		//http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	type IPstatus struct {
		IP     string
		Banned bool
	}

	switch r.Method {
	case "GET":
		var remoteIP string
		if val, ok := r.Header["X-Real-Ip"]; ok {
			remoteIP = val[0]
		} else {
			remoteIP = strings.Split(r.RemoteAddr, ":")[0]
		}
		isBanned, err := f2bclient.IsIPBanned(remoteIP)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
		s := IPstatus{IP: remoteIP, Banned: isBanned}
		t, _ := template.New("unban").Parse(tmpl)
		t.Execute(w, s)
	case "POST":
		r.ParseForm()
		err := f2bclient.UnbanIP(r.Form["ip"][0])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", unban)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
