package logwriter

import (
	"net/http"
	"log"
	"fmt"
	"sync"
	"io/ioutil"
)

var _gwUrlString string

func Initialize(gwUrlString string) {
	_gwUrlString = gwUrlString
}

func LoadSysinfo(gwUrl string, wg *sync.WaitGroup) error {
	if wg != nil {
		defer wg.Done()		
	}
	log.Printf("Fetching %s...", gwUrl)
	res, err := http.Get(gwUrl)
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

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body.  %s", err)
		return err
	}

	parseBody(buf)
	return nil
}

func parseBody(buf []byte) {

}
