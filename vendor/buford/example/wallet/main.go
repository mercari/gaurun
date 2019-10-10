package main

import (
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/RobotsAndPencils/buford/certificate"
	"github.com/RobotsAndPencils/buford/pushpackage"
)

func loadWWDR(name string) (*x509.Certificate, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(b)
}

func failIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var filename, password, intermediate string

	flag.StringVar(&filename, "c", "", "Path to p12 certificate file")
	flag.StringVar(&password, "p", "", "Password for p12 file.")
	flag.StringVar(&intermediate, "i", "", "Path to WWDR intermediate .cer file")
	flag.Parse()

	cert, err := certificate.Load(filename, password)
	failIfError(err)

	wwdr, err := loadWWDR(intermediate)
	failIfError(err)

	f, err := os.Create("Event.pkpass")
	failIfError(err)
	defer f.Close()

	passFiles := []string{
		"pass.json",
		"background.png",
		"background@2x.png",
		"icon.png",
		"icon@2x.png",
		"logo.png",
		"logo@2x.png",
		"thumbnail.png",
		"thumbnail@2x.png",
	}

	pkg := pushpackage.New(f)
	for _, name := range passFiles {
		pkg.File(name, "./Event.pass/"+name)
	}

	err = pkg.Sign(cert, wwdr)
	failIfError(err)
}
