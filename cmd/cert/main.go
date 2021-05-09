package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

func main() {
	serverPem, err := ioutil.ReadFile("server.pem")
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		return
	}

	block, _ := pem.Decode(serverPem)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		fmt.Printf("Err: %s\n", err)
		return
	}

	fmt.Printf("pem data: notBefore=%s - notAfter=%s\n", cert.NotBefore, cert.NotAfter)
}
