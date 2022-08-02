package main

import (
	"time"
)

type AcsChanim struct {
	Wl0 *map[string]map[string]interface{} `json:"wl0,omitempty"`
	Wl1 *map[string]map[string]interface{} `json:"wl1,omitempty"`
}

type AcsChanimStatic struct {
	Badplcp   uint32
	Badtx     uint32
	Doze      uint32
	Glitch    uint32
	Goodtx    uint32
	Inbss     uint32
	Nocat     uint32
	Nopkt     uint32
	Obss      uint32
	Tx        uint32
	Txop      uint32
	Knoise    int32
	Timestamp int
	Chanspec  string
}

type AcsCscore struct {
	Wl0 *map[string]map[string]interface{} `json:"wl0,omitempty"`
	Wl1 *map[string]map[string]interface{} `json:"wl1,omitempty"`
}

type AcsBSSCounters struct {
	Wl0 *map[string]map[string]interface{} `json:"wl0,omitempty"`
	Wl1 *map[string]map[string]interface{} `json:"wl1,omitempty"`
}

type AcsInfo struct {
	Wl0 *map[string]map[string]interface{} `json:"wl0,omitempty"`
	Wl1 *map[string]map[string]interface{} `json:"wl1,omitempty"`
}

type AcsDfsReentry struct {
	Wl0 *map[string]map[string]interface{} `json:"wl0,omitempty"`
	Wl1 *map[string]map[string]interface{} `json:"wl1,omitempty"`
}

type AcsRecords struct {
	Wl0 *map[string]map[string]interface{} `json:"wl0,omitempty"`
	Wl1 *map[string]map[string]interface{} `json:"wl1,omitempty"`
}

type WifiData struct {
	Header     header
	AcsData    *AcsData
	deviceInfo *deviceInfo
}

type AcsData struct {
	CScanPerformed bool
	ACSPerformed   bool
	Info           *AcsInfo        `json:"Info,omitempty"`
	Chanim         *AcsChanim      `json:"Chanim,omitempty"`
	Cscore         *AcsCscore      `json:"Cscore,omitempty"`
	BSSCounters    *AcsBSSCounters `json:"BSSCounters,omitempty"`
	AcsRecords     *AcsRecords     `json:"AcsRecords,omitempty"`
	DfsReentry     *AcsDfsReentry  `json:"DfsReentry,omitempty"`
}

type header struct {
	Timestamp            uint64
	NetworkIdentifier    string
	NetworkIdentifierMac string
}

type config struct {
	postUrl         string
	sampleInterval  time.Duration
	performCScan    bool
	ssl             bool
	compressionType string
}

type deviceInfo struct {
	Model            string
	Interfaces       [2]string
	EnabledDumpTypes map[string]struct{}
	SerialNumber     string
	MacAddress       string
	Executables      map[string]string // acs_cli -> /usr/local/bin/acs_cli ... etc
}
