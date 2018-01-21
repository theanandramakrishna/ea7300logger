package logwriter

import (
	"sync"
	"log"
)

var _inchan chan LogData
var _logdataTable map[string]([]LogData)

func initProcessor(inchan chan LogData) {
	_inchan = inchan

	_logdataTable = make(map[string][]LogData)
}

func process(wg *sync.WaitGroup) {
	defer wg.Done()

	for ; ; {
		var data LogData = <- _inchan

		table := _logdataTable[data.SrcIP]
		if table == nil {
			addData(data)
			continue
		}
		
		if !findValue(data, table) {
			addData(data)
		}
	}
}

func findValue(data LogData, table []LogData) bool {
	for _, val := range table {
		if val == data {
			return true
		}
	}

	return false
}

func addData(data LogData) {
	log.Printf("-------------- time %s|FROM %s|TO %s", data.TimeStamp, data.SrcIP, data.DestIP)
	_logdataTable[data.SrcIP] = append(_logdataTable[data.SrcIP], data)
}



