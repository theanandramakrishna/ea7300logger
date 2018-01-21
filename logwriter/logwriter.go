package logwriter

import (
	"net/http"
	"log"
	"fmt"
	"sync"
	"io"
	"strings"
	"bufio"
	"time"
)

const TAIL_CMD = "tail -200 /var/log/messages"
const IPV6_CMD = "/var/log/ipv6.log"
const UTOPIA = "UTOPIA: FW.LAN2WAN ACCEPT"
const NONE = "(none)"
const SRC = "SRC"
const DST = "DST"

var _gwUrlString string
var _httpClient http.Client = http.Client {}
var _username string
var _password string
var _filtersrc []string
var _outchan chan LogData

type LogData struct {
	TimeStamp time.Time
	SrcIP string
	DestIP string
}

func Initialize(gwUrlString string, username string, password string, filtersrc []string, outchan chan LogData) {
	_gwUrlString = gwUrlString
	_username = username
	_password = password
	_filtersrc = filtersrc
	_outchan = outchan
}

func Start(gwUrl string, wg *sync.WaitGroup) error {
	defer wg.Done()		

	initProcessor(_outchan)
	wg.Add(1)
	go process(wg)

	for ; ; {
		log.Printf("Fetching %s...", gwUrl)
		err := doRequest(gwUrl)
		if err != nil {
			return err
		}
		log.Printf("Sleeping 5 seconds...")
		time.Sleep(time.Second * 5)
	}
}
func doRequest(gwUrl string) error {
	req, err := http.NewRequest("GET", gwUrl, nil)
	if err != nil {
		log.Printf("Could not create request. %s", err)
		return err
	}

	req.SetBasicAuth(_username, _password)

	res, err := _httpClient.Do(req)
	if err != nil {
		log.Printf("Could not connect to %s with supplied username and password.", gwUrl)
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Could not connect to %s, received status %d", gwUrl, res.StatusCode)
		log.Print(err)
		return err
	}

	parseBody(res.Body)	
	return nil
}

func parseBody(body io.Reader) {
	// Assume utf8 and only english characters for now

	bufBody := bufio.NewReader(body)
	for ; ; {
		line, err := bufBody.ReadString('\n')
		if len(line) > 0 {
			// look for the tail command
			if strings.Index(line, TAIL_CMD) != -1 {
				// FOUND!  parse the messages
				parseVarLogMessages(bufBody)
				break
			}
		}
		if err != nil {
			break
		}
	}
}

func parseVarLogMessages(bufBody *bufio.Reader) {
	for ; ; {
		line, err := bufBody.ReadString('\n')
		if len(line) > 0 {
			if strings.Index(line, IPV6_CMD) != -1 {
				// Got to end of var log messages
				break
			}
			logData := parseLogLine(line)
			if logData == nil {
				continue
			}

			outputLogData(logData)
		}
		if err != nil {
			break
		}
	}
}

func parseLogLine(line string) *LogData {
	var logData LogData
	var err error

	idx := strings.Index(line, UTOPIA)
	if idx == -1 {
		// No match
		return nil
	}

	var lineVals []string = strings.Split(strings.TrimSpace(line[idx + len(UTOPIA):]), " ")
	if len(lineVals) == 0 {
		return nil
	}
	
	idx = strings.Index(line, NONE)
	if idx == -1 {
		return nil
	}
	
	logData.TimeStamp, err = time.Parse(time.Stamp, strings.TrimSpace(line[:idx]))
	if err != nil {
		log.Printf(err.Error())
		return nil
	}
	
	for _, val := range lineVals {
		var keyval []string = strings.Split(val, "=")
		// keyval[0] should be the name and keyval[1] should be the value
		if len(keyval) != 2 {
			continue
		}

		if keyval[0] == SRC {
			logData.SrcIP = keyval[1]
		} else if keyval[0] == DST {
			logData.DestIP = keyval[1]
		}
	}

	for _, val := range _filtersrc {
		if logData.SrcIP != val {
			return nil;	// src ip address not found in filter
		}
	}
	return &logData
	
}

func outputLogData(logData *LogData) {
	//log.Printf("++++++++++ time %s|FROM %s|TO %s", logData.TimeStamp, logData.SrcIP, logData.DestIP)
	_outchan <- *logData
}