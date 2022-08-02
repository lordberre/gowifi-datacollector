package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lordberre/gowifi-datacollector/pkg/queue"
	log "github.com/sirupsen/logrus"
)

func getAcsData(cfg *config, deviceInfo *deviceInfo) (*AcsData, error) {
	acs := &AcsData{
		Chanim: &AcsChanim{},
		Cscore: &AcsCscore{},
	}
	for _, wlInterface := range deviceInfo.Interfaces {

		/*/
		if cfg.performCScan {
			// ...
		}
		/*/
		if _, ok := deviceInfo.EnabledDumpTypes["chanim"]; ok {
			if err := acs.Chanim.getChanimDump(wlInterface, deviceInfo.Executables["acs_cli"]); err != nil {
				log.Errorf("[getAcsData] %s", err)

			}
		}
		if _, ok := deviceInfo.EnabledDumpTypes["cscore"]; ok {
			if err := acs.Cscore.getCSoresDump(wlInterface, deviceInfo.Executables["acs_cli"]); err != nil {
				log.Errorf("[getAcsData] %s", err)
			}
		}
	}
	return acs, nil
}

func postHttpData(client *http.Client, url string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Close = true
	req.Header.Set("User-Agent", "Go_WiFi-datacollector/1.0")
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Errorf("Error creating request: %s", err)
	}
	return client.Do(req)
}

/*/
func (w *WifiData) deDup() chan WifiData {
	ch := make(chan WifiData)
	var current WifiData
	go func() {
		for {
			msg, ok := <-ch
			if !ok {
				break
			}
			if w.AcsData.Chanim != nil && previousData.AcsData.Chanim != nil {
				if w.AcsData.Chanim.Wl0 != nil && previousData.AcsData.Chanim.Wl0 != nil {
					if w.AcsData.Chanim.Wl0["timestamp"] == previousData.AcsData.Chanim.Wl0["timestamp"] {
					}
				}
				if w.AcsData.Chanim.Wl1 != nil && previousData.AcsData.Chanim.Wl1 != nil {

				}
			}
		}
	}()
	return ch
}
/*/

func httpPublisher(cfg *config) chan WifiData {
	ch := make(chan WifiData, 10)
	queue := queue.NewQueue()
	client := &http.Client{Timeout: 30 * time.Second}

	go func() {
		for {
			msg, ok := <-ch
			if !ok {
				break
			}
			jsonData, err := json.Marshal(msg)
			if err != nil {
				log.Errorf("Json encode error: %s\n", err)
				continue
			}

			resp, err := postHttpData(client, cfg.postUrl, jsonData)
			if err != nil {
				log.Errorf("Error publishing data: %s. Queueing item.", err)
				if !queue.Push(jsonData) {
					log.Errorf("Queue is full. Replacing oldest item..")
					queue.Pop()
					queue.Push(jsonData)
				}
				log.Infof("Pushed data to queue: %s\n", string(jsonData))
			} else {
				// Do one more attempt for each queue item after a successful post occured
				for {
					if queue.GetLength() == 0 {
						break
					}
					fmt.Printf("Working on %d queued items.\n", queue.GetLength())
					jsonData, _ := queue.Pop()
					log.Infof("Popped data from the queue: %s\n", string(jsonData))
					resp, _ = postHttpData(client, cfg.postUrl, jsonData)
				}
			}
			if resp != nil {
				resp.Body.Close()
			}
			queue.WritePersistent()
		}
	}()
	return ch
}
