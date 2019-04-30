package gateway

import (
	"log"
	"net/http"
	"testing"
)

func helloServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("[]"))
	// fmt.Fprintf(w, "This is an example server.\n")
	// io.WriteString(w, "This is an example server.\n")
}

func startServer() {
	http.HandleFunc("/tyk/apis/", helloServer)
	err := http.ListenAndServeTLS(":7878", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func TestUnsignedCert(t *testing.T) {
	go startServer()

	x, err := NewGatewayClient("https://localhost:7878", "")
	if err != nil {
		t.Fatal(err)
	}
	x.SetInsecureTLS(true)

	_, err = x.FetchAPIs()
	if err != nil {
		t.Fatal(err)
	}

}
