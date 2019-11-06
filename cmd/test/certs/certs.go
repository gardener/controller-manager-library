package certs

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gardener/controller-manager-library/pkg/certmgmt"
	"github.com/gardener/controller-manager-library/pkg/certmgmt/secret"
	"github.com/gardener/controller-manager-library/pkg/certs/access"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"net/http"
	"time"
)

func CertsMain() {

	cfg := &certmgmt.Config{
		CommonName: "test",
		DnsNames:   []string{"test.mandelsoft.org"},
		Rest:       24 * time.Hour,
		Validity:   7 * 24 * time.Hour,
	}
	i := certmgmt.NewCertInfo(nil, nil, nil, nil)
	n, err := certmgmt.UpdateCertificate(i, cfg)
	if err != nil {
		fmt.Printf("Initial creation failed: %s", err)
		return
	}

	if !certmgmt.IsValid(n, "test.mandelsoft.org", 24*time.Hour) {
		fmt.Printf("not valid for 24h")
		return
	}

	if certmgmt.IsValid(n, "test.mandelsoft.org", cfg.Validity) {
		fmt.Printf("valid for more than initial validity")
		return
	}

	if !certmgmt.IsValid(n, "", 24*time.Hour) {
		fmt.Printf("not valid for no dnsnames")
		return
	}

	r, err := certmgmt.UpdateCertificate(n, cfg)
	if err != nil {
		fmt.Printf("update failed: %s", err)
		return
	}
	if !certmgmt.IsValid(r, "test.mandelsoft.org", 24*time.Hour) {
		fmt.Printf("not valid for 24h")
		return
	}

	c, err := cluster.CreateCluster(context.Background(), logger.New(), cluster.Configure("dummy", "", "").Definition(), "", "")
	if err != nil {
		fmt.Printf("no cluster: %s\n", err)
		return
	}

	fmt.Printf("********************\n")
	secret := secret.NewSecret(c, resources.NewObjectName("default", "access"))

	cfg.DnsNames = []string{"localhost"}

	fmt.Printf("setting up certificate watch\n")
	cert, err := access.New(context.TODO(), logger.New(), secret, cfg)
	if err != nil {
		fmt.Printf("get certmgmt failed: %s\n", err)
		return
	}

	tlscfg := &tls.Config{
		GetCertificate: cert.GetCertificate,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", HelloServer)
	server := http.Server{
		Addr:      ":4443",
		TLSConfig: tlscfg,
		Handler:   mux,
	}

	fmt.Printf("Starting server\n")
	server.ListenAndServeTLS("", "")
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
	// fmt.Fprintf(w, "This is an example server.\n")
	// io.WriteString(w, "This is an example server.\n")
}
