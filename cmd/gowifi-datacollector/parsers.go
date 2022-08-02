package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	chanimExpectedColumns  = 14
	cScoresExpectedColumns = 18
)

/*/
Dumps and parses chanim values from acs_cli (acsd version: 2)
Example:
0 = Chanim Stats: version: 4, count: 19
1 = chanspec tx   inbss   obss   nocat   nopkt   doze   txop   goodtx  badtx  glitch   badplcp  knoise  timestamp
2 = 0xd024      0       0       1       0       1       0       96      0       0       24      0       -97     981738759
3 = (....)
/*/
func (c *AcsChanim) getChanimDump(wlInterface string, executable string) error {
	var chanimErrors error

	// Fetch data from ACS daemon
	// cmd := exec.Command("/usr/sbin/acs_cli", "-i", wlInterface, "dump", "chanim")
	cmd := exec.Command(executable, "-i", wlInterface, "dump", "chanim")
	data, err := cmd.Output()
	if err != nil {
		return err
	}

	// Parse output
	channels := strings.Split(string(data), "\n")
	if len(channels) <= 1 {
		chanimErrors = fmt.Errorf("invalid chanim data for %s", wlInterface)
		return chanimErrors
	}
	var columns []string
	chanimMap := make(map[string]map[string]interface{})
	if wlInterface == "wl0" {
		c.Wl0 = &chanimMap
	} else if wlInterface == "wl1" {
		c.Wl1 = &chanimMap
	}
	for rowNum, data := range channels {

		// Skip the header
		if rowNum == 0 {
			continue
		}

		var chanspec string
		fields := strings.Fields(data)
		if len(fields) != chanimExpectedColumns {
			continue
		}
		// The columns names
		if rowNum == 1 {
			columns = fields
		} else if rowNum >= 2 { // The data
			for i, col := range columns {
				if i == 0 {
					chanspec = chanSpecs[fields[i]]
				}

				// Try to convert it to numeric
				var value interface{}
				if v, err := strconv.ParseInt(fields[i], 10, 32); err != nil {
					value = fields[i]
				} else {
					value = int32(v)
				}
				if _, ok := chanimMap[chanspec]; !ok {
					chanimMap[chanspec] = map[string]interface{}{col: value}
				} else {
					chanimMap[chanspec][col] = value
				}
			}
		}
	}
	return nil
}

/*/
Dumps and parses cscore values from acs_cli (acsd version: 2)
Example:
ACSD Candidate Scores for next Channel Switch:
   Channel (Chspec) Use DFS      BSS     busy  interf.  itf_adj      fcs  txpower  bgnoise    TOTAL      CNS      ADJ     TXOP  DFS/Age     RBSS    RTXOP     CHWT
36/80      (0xe02a)   -        50        0        1        0        0       20        0      231      -96        5      150       10        0        0        0
40/80     (.....)
/*/
func (c *AcsCscore) getCSoresDump(wlInterface string, executable string) error {

	// Fetch data from ACS daemon
	// cmd := exec.Command("/usr/sbin/acs_cli", "-i", wlInterface, "dump", "cscore")
	cmd := exec.Command(executable, "-i", wlInterface, "dump", "cscore")
	data, err := cmd.Output()
	if err != nil {
		return err
	}

	// Parse output
	channels := strings.Split(string(data), "\n")
	if len(channels) <= 1 {
		return fmt.Errorf("invalid cScores data for %s. Reason: Empty or partially empty dump response", wlInterface)
	}
	var columns []string
	cScoresMap := make(map[string]map[string]interface{})
	if wlInterface == "wl0" {
		c.Wl0 = &cScoresMap
	} else if wlInterface == "wl1" {
		c.Wl1 = &cScoresMap
	}
	for rowNum, data := range channels {

		// Skip the header
		if rowNum == 0 {
			continue
		}

		fields := strings.Fields(data)
		// The columns names
		if rowNum == 1 {

			if len(fields) != cScoresExpectedColumns+1 {
				return fmt.Errorf("invalid cScores data for %s. Reason: %d Not maching cScoresExpectedColumns: %d",
					wlInterface, len(fields), cScoresExpectedColumns)
			}

			columns = make([]string, 0)
			for i := range fields {
				if i == 1 { // Clean chanspec column name
					chanspec := fields[i]
					for _, stringCleanup := range []string{"(", ")"} {
						chanspec = strings.Replace(chanspec, stringCleanup, "", 1)
					}
					fields[i] = chanspec
				}
				// Fix the whitespace in "Use DFS" column so column count matches.
				if i == 2 {
					col := fields[i] + fields[i+1]
					columns = append(columns, col)
				} else if i == 3 {
					continue
				} else {
					columns = append(columns, fields[i])
				}
			}
		} else if rowNum >= 2 { // The data
			if len(fields) != cScoresExpectedColumns {
				continue
			}
			var (
				chanspec    string
				chanSpecHex string
			)
			for i, col := range columns {
				if i == 0 {
					chanspec = fields[1]
					for _, stringCleanup := range []string{"(", ")"} {
						chanspec = strings.Replace(chanspec, stringCleanup, "", 1)
					}
					chanSpecHex = chanspec
					chanspec = chanSpecs[chanspec]
				}

				var value interface{}

				// Dont convert chanspec, channel columns
				if i == 1 {
					value = chanSpecHex
				} else if i == 0 {
					value = fields[i]
				} else if i == 2 { // Convert UseDFS ("dfs", "-") to boolean
					if fields[i] == "dfs" {
						value = true
					} else {
						value = false
					}
				} else {
					// Try to convert it to numeric
					if v, err := strconv.ParseInt(fields[i], 10, 32); err != nil {
						value = fields[i]
					} else {
						value = int32(v)
					}
				}
				if _, ok := cScoresMap[chanspec]; !ok {
					cScoresMap[chanspec] = map[string]interface{}{col: value}
				} else {
					cScoresMap[chanspec][col] = value
				}
			}
		}
	}
	return nil
}

func (d *deviceInfo) getChanSpecs() error {
	chanSpecs = make(map[string]string)
	for _, interFace := range d.Interfaces {
		cmd := exec.Command(d.Executables["wl"], "-i", interFace, "chanspecs")
		data, err := cmd.Output()
		if err != nil {
			return err
		}
		chanspecs := strings.Split(string(data), "\n")
		if len(chanspecs) <= 1 {
			return fmt.Errorf("invalid chanspec data for %s. Reason: Empty or partially empty dump response", interFace)
		}
		for _, chanspec := range chanspecs {
			rows := strings.Fields(chanspec)
			if len(rows) == 0 {
				continue
			}
			// 0xe02a --> 36/80
			chanSpecReadable := rows[0]
			chanSpecHex := rows[1]
			for _, stringCleanup := range []string{"(", ")"} {
				chanSpecHex = strings.Replace(chanSpecHex, stringCleanup, "", 1)
			}
			chanSpecs[chanSpecHex] = chanSpecReadable
		}
	}
	return nil
}

func (d *deviceInfo) getOperatingInterfaces() error {
	for i, interFace := range [2]string{"wl0", "wl1"} {
		cmd := exec.Command(d.Executables["wl"], "-i", interFace, "isup")
		data, err := cmd.Output()
		if err != nil {
			return err
		}
		if state, err := strconv.Atoi(strings.Trim(string(data), "\n")); err == nil {
			if state == 1 {
				d.Interfaces[i] = interFace
			}
		} else {
			return err
		}
	}
	fmt.Printf("Detected these operating interfaces: %+v\n", d.Interfaces)
	return nil
}
