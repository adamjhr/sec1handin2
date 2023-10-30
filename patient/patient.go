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
	"math/rand"
	"net/http"

	"github.com/adamjhr/sec1handin2/types"
	"github.com/adamjhr/sec1handin2/util"
)

var client *http.Client
var data int
var p int
var hospitalPort int
var port int
var dataMax int
var expectedShares int
var receivedShares []int

// List of ports of other patients provided by hospital
func patients(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		log.Println(port, ": Patient received POST /patients")

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(port, ": Error when reading request body during /patients:", err)
			w.WriteHeader(418)
			return
		}
		patients := &types.Patients{}
		err = json.Unmarshal(body, patients)
		if err != nil {
			log.Fatal(port, ": Error when unmarshalling patients:", err)
			w.WriteHeader(418)
			return
		}
		expectedShares = len(patients.PortsList) + 1 // Expected amount of shares in the receivedShares-list

		shares := util.CreateShares(p, data, expectedShares) // Create shares from own data

		log.Println(port, ": Sending shares to other patients")
		for i, share := range shares { // Send a share to each other patient
			ownShare := types.Share{
				Share: share,
			}
			b, err := json.Marshal(ownShare)
			if err != nil {
				log.Fatal(port, ": Error when marshalling share during /patients:", err)
				w.WriteHeader(418)
				return
			}
			url := fmt.Sprintf("https://localhost:%d/shares", patients.PortsList[i])
			response, err := client.Post(url, "string", bytes.NewReader(b))
			if err != nil {
				log.Fatal(port, ": Error when sending share to", patients.PortsList[i], ":", err)
				w.WriteHeader(418)
				return
			}
			log.Println(port, ": Sent share to, ", patients.PortsList[i], ". Received response code:", response.StatusCode)
		}

		receivedShares = append(receivedShares, shares[len(shares)-1]) // Add own share to list of receivedShares

		w.WriteHeader(200)

	}
}

func shares(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		log.Println(port, ": Patient received POST /shares")

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(port, ": Error when reading request body during /shares:", err)
			w.WriteHeader(418)
			return
		}
		foreignShare := &types.Share{}
		err = json.Unmarshal(body, foreignShare)
		if err != nil {
			log.Fatal(port, ": Error when unmarshalling share:", err)
			w.WriteHeader(418)
			return
		}

		receivedShares = append(receivedShares, foreignShare.Share) // Add received share to list of received shares

		if len(receivedShares) == expectedShares-1 {
			log.Println(port, ": Computing aggregate share")
			// If all shares have been received,
			// calculate and aggregate share and send it to the hospital

			var aggregateShare int

			for _, share := range receivedShares {
				aggregateShare = aggregateShare + share
			}

			aggregateShare = aggregateShare % p

			aggregate := types.Share{
				Share: aggregateShare,
			}

			b, err := json.Marshal(aggregate)
			if err != nil {
				log.Fatal(port, ": Error when marshalling aggregate share during /shares:", err)
				w.WriteHeader(418)
				return
			}

			log.Println(port, ": Sending aggregate share", aggregateShare, "to hospital")
			url := fmt.Sprintf("https://localhost:%d/shares", hospitalPort)
			response, err := client.Post(url, "string", bytes.NewReader(b))
			if err != nil {
				log.Fatal(port, ": Error when sending aggregate share to hospital:", err)
				w.WriteHeader(418)
				return
			}
			log.Println(port, ": Sent aggregate share to hospital, received response code", response.StatusCode)

		}

	}

}

func patientServer() {
	log.Println(port, ": Creating patient server")

	mux := http.NewServeMux()
	mux.HandleFunc("/patients", patients)
	mux.HandleFunc("/shares", shares)

	err := http.ListenAndServeTLS(util.StringifyPort(port), "server.crt", "server.key", mux)
	if err != nil {
		log.Fatal(err)
	}

}

func Main() {

	flag.IntVar(&port, "port", 8081, "port for patient")
	flag.IntVar(&hospitalPort, "h", 8080, "port of the hospital")
	flag.IntVar(&p, "p", 500, "order of the cyclical group")
	flag.IntVar(&dataMax, "m", 8080, "max value of the patient data")

	data := rand.Intn(dataMax-1) + 1

	log.Println(port, ": New patient with p =", p, "and data =", data)

	// Load in the certification from file server.crt
	cert, err := ioutil.ReadFile("server.crt")
	if err != nil {
		log.Fatal(err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cert)

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
	}

	go patientServer() // Run the server

	// Register patient with hospital
	log.Println(port, ": Patient registering with hospital")

	url := fmt.Sprintf("https://localhost:%d/patient", hospitalPort)

	ownPort := types.Patient{
		Port: port,
	}

	b, err := json.Marshal(ownPort)
	if err != nil {
		log.Fatal(port, ": Error when marshalling patient:", err)
	}

	response, err := client.Post(url, "string", bytes.NewReader(b))
	if err != nil {
		log.Fatal(port, ": Error when regisering with hospital:", err)
	}
	log.Println(port, ": Registered with hospital, received response code", response.Status)

}
