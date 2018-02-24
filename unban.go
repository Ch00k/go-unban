package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func runFail2banClient(args ...string) string {
	binary := "fail2ban-client"
	cmd := exec.Command(binary, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Fatal: %s\n%s", err, stderr.String())
	}
	return stdout.String()
}

func getJails() []string {
	out := runFail2banClient("status")
	parts := strings.Split(out, "Jail list:")
	jails := strings.TrimSpace(parts[1])
	return strings.Split(jails, ", ")
}

func getJailBannedIPs(jail string) []string {
	out := runFail2banClient("status", jail)
	parts := strings.Split(out, "Banned IP list:")
	ips := strings.TrimSpace(parts[1])
	return strings.Split(ips, " ")
}

func isIPBanned(ip string, jails []string) bool {
	for _, jail := range jails {
		ips := getJailBannedIPs(jail)
		if contains(ips, ip) {
			return true
		}
	}
	return false
}

func unbanIP(ip string) {
	runFail2banClient("unban", ip)
}

func contains(array []string, item string) bool {
	for _, i := range array {
		if i == item {
			return true
		}
	}
	return false
}

func unban(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	fmt.Println(r.Header)

	type IPstatus struct {
		IP     string
		Banned bool
	}

	switch r.Method {
	case "GET":
		remoteIP := r.Header["X-Real-Ip"][0]
		jails := getJails()
		isBanned := isIPBanned(remoteIP, jails)
		s := IPstatus{IP: remoteIP, Banned: isBanned}
		t, _ := template.ParseFiles("template.html")
		t.Execute(w, s)
	case "POST":
		r.ParseForm()
		unbanIP(r.Form["ip"][0])
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", unban)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
