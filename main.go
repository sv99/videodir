package main

import (
	"log"
	"fmt"
	"os"
	"io/ioutil"
	"path"
	"github.com/stvp/go-toml-config"
	"github.com/julienschmidt/httprouter"
	"crypto/tls"
	"crypto/x509"
	"net/http"
)

var (
	serverAddr = config.String("serverAddr", ":8080")
	videoDir   = config.String("videoDir", "")
	cert       = config.String("cert", "cert.pem")
	key        = config.String("key", "key.pem")
)

func Index(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "videoDir: %s\n", *videoDir)
}

func main() {
	//log.SetFlags(log.Lshortfile)
	//get path name for the executable
	ex, err := os.Executable()
	if err != nil {
		log.Println(err)
		panic(err)
	}
	exPath := path.Dir(ex)

	//read configuration
	if err := config.Parse(path.Join(exPath, "videodir.conf")); err != nil {
		log.Println(err)
		panic(err)
	}

	log.Println("videoDir: ", *videoDir)
	log.Println("key: ", *key)
	log.Println("cert: ", *cert)

	//run server
	certPath := path.Join(exPath, *cert)
	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		log.Fatalln("Unable to read ", certPath, err)
	}

	myCertPool := x509.NewCertPool()
	if ok := myCertPool.AppendCertsFromPEM(certBytes); !ok {
		log.Fatalln("Unable to add certificate to certificate pool")
	}

	tlsConfig := &tls.Config{
		// Reject any TLS certificate that cannot be validated
		ClientAuth: tls.RequireAndVerifyClientCert,
		// Ensure that we only use our "CA" to validate certificates
		ClientCAs: myCertPool,
		// PFS because we can
		CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384},
		// Force it server side
		PreferServerCipherSuites: true,
		// TLS 1.2 because we can
		MinVersion: tls.VersionTLS12,
	}

	tlsConfig.BuildNameToCertificate()

	router := httprouter.New()
	router.GET("/", Index)
	router.ServeFiles("/video/*filepath", http.Dir(*videoDir))

	httpServer := &http.Server{
		Addr:      *serverAddr,
		Handler:   router,
		TLSConfig: tlsConfig,
	}

	log.Println(httpServer.ListenAndServeTLS(certPath, path.Join(exPath, *key)))

	// Start the HTTPS server in a goroutine
	//fmt.Println(fmt.Sprintf("Start server on https://localhost%s", *serverAddr))
	//go http.ListenAndServeTLS(":8081", "cert.pem", "key.pem", router)
	// Start the HTTP server
	//log.Println(fmt.Sprintf("Start server on http://localhost%s", *serverAddr))
	//http.ListenAndServe(*serverAddr, router)
}
