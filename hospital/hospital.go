package hospital

import (
	"log"
	"net/http"
)

func RegisterPatient(w http.ResponseWriter, req *http.Request) {
	log.Println("Index")
	if req.Method == "GET" {
		w.Write([]byte("Hello"))
	}
}

func Main() {

	http.HandleFunc("/registerpatient", RegisterPatient)
	err := http.ListenAndServeTLS(":8081", "server.crt", "server.key", nil)

	if err != nil {
		log.Fatal(err)
	}

}
