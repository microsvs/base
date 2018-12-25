package utils

import (
	"net"
	"net/http"
	"strings"
)

//GetClientIPAddress from http.Request
func GetClientIPAdress(r *http.Request) string {
	var ipAddress string
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		for _, ip := range strings.Split(r.Header.Get(h), ",") {
			// header can contain spaces too, strip those out.
			realIP := net.ParseIP(strings.Replace(ip, " ", "", -1))
			ipAddress = string(realIP)
		}
	}
	return ipAddress
}
