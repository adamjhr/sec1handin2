package hospital

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/adamjhr/sec1handin2/types"
	"github.com/adamjhr/sec1handin2/util"
)

var patients []int
var data int

func RegisterPatient(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(418)
		}
		port_string := string(body)
		port, err := strconv.Atoi(port_string)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(418)
		}
		log.Println("Registered new patient at port: ", port)
		patients = append(patients, port)
		w.WriteHeader(200)
	}
}

func AddShare(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(418)
			return
		}
		share := &types.Share{}
		err = json.Unmarshal(body, share)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(418)
			return
		}
		data += share.Share
	}
}

func HospitalServer(port int) {
	http.HandleFunc("/patient", RegisterPatient)
	err := http.ListenAndServeTLS(util.StringifyPort(port), "server.crt", "server.key", nil)

	if err != nil {
		log.Fatal(err)
	}
}

func Hospital(port int) {

	data = 0

	go HospitalServer(port)

	time.Sleep(time.Second)

	cert, err := ioutil.ReadFile("server.crt")
	if err != nil {
		log.Fatal(err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
	}

	for i, p := range patients {
		otherPatients := patients
		otherPatients[i] = otherPatients[len(otherPatients)-1]
		otherPatients = otherPatients[:len(otherPatients)-1]

		url := fmt.Sprintf("https://localhost:%d/patients", p)
		patientPorts := types.Patients{
			PortsList: otherPatients,
		}

		b, err := json.Marshal(patientPorts)
		if err != nil {
			log.Fatal(err)
		}

		response, err := client.Post(url, "application/json", bytes.NewReader(b))
		if err != nil {
			log.Fatal(err)
		}
		log.Println(response.Status)

	}

}

func Main() {
	Hospital(8080)
}
