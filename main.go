package main

import (
	"github.com/adamjhr/sec1handin2/hospital"
	"github.com/adamjhr/sec1handin2/patient"
)

func main() {
	go hospital.Hospital(8080)
	go patient.Patient(8081, 8080, 500)

	for {
	}
}
