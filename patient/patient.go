package patient

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

	"github.com/adamjhr/sec1handin2/types"
	"github.com/adamjhr/sec1handin2/util"
)

var client *http.Client
var data int
var r int
var expectedShares int
var receivedShares []int

func patients(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(418)
			return
		}
		patients := &types.Patients{}
		err = json.Unmarshal(body, patients)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(418)
			return
		}
		expectedShares = len(patients.PortsList) + 1

		shares := util.CreateShares(r, data, expectedShares)

		for i, share := range shares {
			ownShare := types.Share{
				Share: share,
			}
			b, err := json.Marshal(ownShare)
			if err != nil {
				log.Fatal(err)
				w.WriteHeader(418)
				return
			}
			url := fmt.Sprintf("https://localhost:%d/shares", patients.PortsList[i])
			response, err := client.Post(url, "string", bytes.NewReader(b))
			if err != nil {
				log.Fatal(err)
				w.WriteHeader(418)
				return
			}
			log.Println(response.StatusCode)
			body, err := io.ReadAll(req.Body)
			if err != nil {
				log.Fatal(err)
				w.WriteHeader(418)
				return
			}
			foreignShare := &types.Share{}
			err = json.Unmarshal(body, foreignShare)
			if err != nil {
				log.Fatal(err)
				w.WriteHeader(418)
				return
			}
		}

		receivedShares = append(receivedShares, shares[len(shares)-1])

		w.WriteHeader(200)

	}
}

func shares(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
}

func patientServer(port int) {
	http.HandleFunc("/patients", patients)
	http.HandleFunc("/shares", shares)

	err := http.ListenAndServeTLS(util.StringifyPort(port), "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal(err)
	}

}

func Patient(port int, hospitalPort int, r int) {

	log.Println("New patient with port", port, "and r =", r)
	r = r

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

	go patientServer(port)

	portBytes := bytes.NewReader([]byte(fmt.Sprintf("%d", port)))
	url := fmt.Sprintf("https://localhost:%d/patient", hospitalPort)
	response, err := client.Post(url, "string", portBytes)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(response.Status)

}

func Main() {
	Patient(8081, 8080, 500)
}
