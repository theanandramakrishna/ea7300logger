package main

import (
	"fmt"
	"flag"
	"errors"
	"github.com/theanandramakrishna/ea7300logger/logwriter"
	"sync"
)

const GATEWAY_HELP = "ip address in x.x.x.x form for the gateway"
const ADMINUSER_HELP = "admin user name"
const PASSWORD_HELP = "admin password"
const HELP_HELP = "help"

var ipString string
var adminUsername string
var password string
var helpMode bool
var args []string

func main() {
	var err error = initArgs()
	if err != nil {
		printUsage()
		panic(err)
	}

	if helpMode == true {
		printUsage()
		return
	}

	// Have arguments, attempt connect.
	var wg sync.WaitGroup
	var outchan chan logwriter.LogData = make(chan logwriter.LogData, 10)

	var gwUrlString = fmt.Sprintf("http://%s/sysinfo.cgi", ipString)
	logwriter.Initialize(gwUrlString, adminUsername, password, args, outchan)

	wg.Add(1)	
	go logwriter.Start(gwUrlString, &wg)

	wg.Wait()
}

func printUsage() {
	fmt.Printf("Usage: \n\n")
	fmt.Printf("\tgetEA7300Log -g <gateway ip> -u <admin username> -p <password> [ip1] [ip2] ...\n")
}
func initArgs() error {
	// Extract out ip address and password 
	flag.StringVar(&ipString, "gateway", "", GATEWAY_HELP)
	flag.StringVar(&ipString, "g", "", GATEWAY_HELP)
	flag.StringVar(&adminUsername, "adminuser", "", ADMINUSER_HELP)
	flag.StringVar(&adminUsername, "u", "", ADMINUSER_HELP)
	flag.StringVar(&password, "password", "", PASSWORD_HELP)
	flag.StringVar(&password, "p", "", PASSWORD_HELP)
	flag.BoolVar(&helpMode, "help", false, HELP_HELP)
	flag.BoolVar(&helpMode, "h", false, HELP_HELP)

	flag.Parse()

	// Validate args
	if ipString == "" {
		return errors.New("Missing required argument: ipString")
	}

	if adminUsername == "" {
		return errors.New("Missing required argument: admin user name")
		
	}

	args = flag.Args()

	return nil
}

