package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"time"
)

const DeviceInfoInterval = 3600

var (
	version           string
	loglevel          uint
	postUrl           string
	sampleIntervalSec uint
	performCScan      bool
	ssl               bool
	compressiontype   string
)

// start-stop-daemon -S -c datacollector -x /root/gowifi-datacollector
func main() {

	flag.StringVar(&postUrl, "u", "http://127.0.0.1:8080", "URL to post JSON data")
	flag.UintVar(&loglevel, "l", 1, "Set log level (1(warning)/2(info)/3(debug)")
	flag.UintVar(&sampleIntervalSec, "i", 5, "How often in seconds to perform data collection (Minimum 5 sec)")
	performCScan = *flag.Bool("c", false, "Perform a CScan before performing each sample (default=false)")
	printVersion := flag.Bool("v", false, "print version")

	flag.Parse()
	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if sampleIntervalSec < 5 {
		sampleIntervalSec = 5
	}

	cfg := &config{
		postUrl:         postUrl,
		sampleInterval:  time.Duration(sampleIntervalSec),
		performCScan:    performCScan,
		ssl:             ssl,
		compressionType: compressiontype,
	}

	if loglevel == 1 {
		log.SetLevel(log.WarnLevel)
	} else if loglevel == 2 {
		log.SetLevel(log.InfoLevel)
	} else if loglevel == 3 {
		log.SetLevel(log.DebugLevel)
	} else if loglevel == 4 {
		log.SetLevel(log.TraceLevel)
	}

	fmt.Printf("Starting Go Wi-Fi datacollector v%s\n", version)

	output := httpPublisher(cfg)
	/*/
	output := make(chan WifiData, 10)
	go func() {
		for {
			if _, ok := <-output; !ok {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()
	/*/

	for {

		// Get device compatibility
		updateDeviceInfo := time.NewTimer(60 * time.Second)
		deviceInfo, err := getCompatibleDevice()
		if err != nil {
			log.Panicf("Fatal compatibility or dependency error: %s\n", err)
		}

	dataCollectorLoop:
		for {
			select {
			// Update device compatibility every $DeviceInfoInterval
			case <-updateDeviceInfo.C:
				fmt.Printf("Updating deviceInfo...\n")
				break dataCollectorLoop
			default:
				telemetryEnvelope := &WifiData{
					Header: header{
						Timestamp:            uint64(time.Now().Unix()),
						NetworkIdentifier:    deviceInfo.SerialNumber,
						NetworkIdentifierMac: deviceInfo.MacAddress,
					},
					deviceInfo: deviceInfo,
				}

				telemetryEnvelope.AcsData, err = getAcsData(cfg, deviceInfo)
				if err != nil {
					log.Errorf("Acs data error: %s\n", err)
					continue
				}
				output <- *telemetryEnvelope
				time.Sleep(cfg.sampleInterval * time.Second)
			}
		}
	}
}
