package certs

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/cert"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/certmgmt"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"net/http"
	"time"
)

func CertsMain() {

	i := cert.NewCertInfo(nil, nil, nil, nil)
	d := 25 * 366 * time.Hour
	n, err := cert.UpdateCertificate(i, "test", "test.mandelsoft.org", d)
	if err != nil {
		fmt.Printf("Initial creation failed: %s", err)
		return
	}

	if !cert.IsValid(n, "test.mandelsoft.org", 24*time.Hour) {
		fmt.Printf("not valid for 24h")
		return
	}

	if cert.IsValid(n, "test.mandelsoft.org", d) {
		fmt.Printf("valid for more than 365 days")
		return
	}

	if !cert.IsValid(n, "", 24*time.Hour) {
		fmt.Printf("not valid for no dnsnames")
		return
	}

	r, err := cert.UpdateCertificate(n, "test", "test.mandelsoft.org", d)
	if err != nil {
		fmt.Printf("update failed: %s", err)
		return
	}
	if !cert.IsValid(r, "test.mandelsoft.org", 24*time.Hour) {
		fmt.Printf("not valid for 24h")
		return
	}

	c, err := cluster.CreateCluster(context.Background(), logger.New(), cluster.Configure("dummy", "", "").Definition(), "", "")
	if err != nil {
		fmt.Printf("no cluster: %s\n", err)
		return
	}

	fmt.Printf("********************\n")
	secret := certmgmt.NewSecret(c, resources.NewObjectName("default", "secret"))

	r, err = certmgmt.GetCertificateInfo(logger.New(), secret, "test", "localhost")
	if err != nil {
		fmt.Printf("get cert failed: %s\n", err)
		return
	}
	cert, err := tls.X509KeyPair(r.Cert(), r.Key())
	if err != nil {
		fmt.Printf("key pair failed: %s", err)
		return
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", HelloServer)
	server := http.Server{
		Addr:      ":4443",
		TLSConfig: cfg,
		Handler:   mux,
	}

	server.ListenAndServeTLS("", "")
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
	// fmt.Fprintf(w, "This is an example server.\n")
	// io.WriteString(w, "This is an example server.\n")
}
