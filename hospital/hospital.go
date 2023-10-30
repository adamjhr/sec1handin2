package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var patients []int
var data int
var p int
var port int
var expectedShares int
var receivedShares int

func Patients(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		log.Println(port, ": Hospital received POST /patients")

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(port, ": Error when reading request body during /patients:", err)
			w.WriteHeader(418)
			return
		}
		receivedPort := &types.Patient{}
		err = json.Unmarshal(body, receivedPort)
		if err != nil {
			log.Fatal(port, ": Error when unmarshalling port:", err)
			w.WriteHeader(418)
			return
		}
		log.Println(port, ": Registered new patient at port", receivedPort.Port)
		patients = append(patients, receivedPort.Port)
		expectedShares++
		w.WriteHeader(200)
	}
}

func Shares(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		log.Println(port, ": Hospital received POST /shares")

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(port, ": Error when reading request body during /shares:", err)
			w.WriteHeader(418)
			return
		}
		share := &types.Share{}
		err = json.Unmarshal(body, share)
		if err != nil {
			log.Fatal(port, ": Error when unmarshalling share:", err)
			w.WriteHeader(418)
			return
		}
		log.Println(port, ": Hospital received share", share.Share)
		data = (data + share.Share) % p
		receivedShares++

		if receivedShares == expectedShares {
			log.Println("Computation finished: The final value is", data)
		}
		w.WriteHeader(200)
	}
}

func HospitalServer() {
	log.Println(port, ": Creating hospital server")

	mux := http.NewServeMux()
	mux.HandleFunc("/patient", Patients)
	mux.HandleFunc("/shares", Shares)

	err := http.ListenAndServeTLS(util.StringifyPort(port), "server.crt", "server.key", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func Main() {

	flag.IntVar(&port, "port", 8080, "port of hospital")
	flag.IntVar(&p, "p", 500, "order of cyclical group")

	data = 0

	go HospitalServer()

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

	log.Println(port, ": Sending ports to patients")
	for i, p := range patients {
		otherPatients := patients
		otherPatients[i] = otherPatients[len(otherPatients)-1]
		otherPatients = otherPatients[:len(otherPatients)-1]

		log.Println(port, ": Sending ports to", p)
		url := fmt.Sprintf("https://localhost:%d/patients", p)
		patientPorts := types.Patients{
			PortsList: otherPatients,
		}

		b, err := json.Marshal(patientPorts)
		if err != nil {
			log.Fatal(port, ": Error when marshalling patientPorts:", err)
		}

		response, err := client.Post(url, "application/json", bytes.NewReader(b))
		if err != nil {
			log.Fatal(port, ": Error when posting patientPorts to", p, ":", err)
		}
		log.Println(port, ": Sent ports to ", p, ". Received response code", response.Status)

	}

}
