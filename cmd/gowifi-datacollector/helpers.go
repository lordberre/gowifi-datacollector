package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Fech and store S/N, Mac address
func (d *deviceInfo) storeDeviceIDs() error {
	var (
		deviceIds = [2]string{"UnknownSN", "UnknownMAC"}
		cmdErrors error
	)
	/*/
	C1:
	mdm getpv Device.DeviceInfo.SerialNumber
	mdm getpv Device.X_BROADCOM_COM_eRouter.FactoryMib.CMIpMacAddress
	/*/
	var (
		paths []string
		args  []string
		cmd   string
	)
	if isSwan() {
		paths = []string{"Device/DeviceInfo/SerialNumber", "Device/DeviceInfo/MACAddress"}
		args = []string{"-gp"}
		cmd = "/usr/bin/xmo-client"
	} else {
		/*/
		paths = []string{"Device.DeviceInfo.SerialNumber", "Device.X_BROADCOM_COM_eRouter.FactoryMib.CMIpMacAddress"}
		args = []string{"getpv"}
		cmd = "function __c() { /bin/echo \"$*\" | /usr/local/bin/consoled; unset -f __c; }; __c"
		/*/
		return errors.New("unsupported device")
	}
	for i, path := range paths {
		pathAndArg := make([]string, 0)
		for _, arg := range args {
			pathAndArg = append(pathAndArg, arg)
			pathAndArg = append(pathAndArg, path)
		}
		result := exec.Command(cmd, pathAndArg...)
		data, err := result.Output()
		if err != nil {
			log.Errorf(err.Error())
			cmdErrors = err
			continue
		}

		if isSwan() {
			// Parse XMO output
			lines := strings.Split(string(data), "\n")
			if len(lines) > 3 {
				cmdErrors = fmt.Errorf("invalid deviceInfo. Result: %s", string(data))
			}
			value := strings.Fields(lines[1])[2]
			deviceIds[i] = value
		} else {
			// Parse MDM output
			lines := strings.Split(string(data), "\n")
			if len(lines) != 9 {
				cmdErrors = fmt.Errorf("invalid deviceInfo. Result: %s", string(data))
			}
			value := strings.Split(lines[2], "=")
			deviceIds[i] = value[1]
		}

	}
	d.SerialNumber = deviceIds[0]
	d.MacAddress = deviceIds[1]
	if d.SerialNumber == "UnknownSN" && d.MacAddress == "UnknownMAC" {
		log.Errorf("[storeDeviceIDs] Couldn't find SN/Mac address for the device")
	}
	return cmdErrors
}

func isSwan() bool {
	cmd := exec.Command("/usr/bin/xmo-client")
	if _, err := cmd.Output(); err == nil {
		return true
	}
	return false
}

func getCompatibleDevice() (*deviceInfo, error) {
	// Check board/product. cat /proc/device-tree/board_id/product_name
	productName := "Unknown"
	cmd := exec.Command("cat", "/proc/device-tree/board_id/product_name")
	name, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	if len(name) <= 1 {
		return nil, errors.New("invalid product name")
	}

	enabledDumpTypes := make(map[string]struct{})
	productName = string(name)
	if productName == "BCM93384WVG\u0000" {
		enabledDumpTypes["chanim"] = struct{}{}
	} else if productName == "3890V3\u0000" {
		enabledDumpTypes["chanim"] = struct{}{}
		enabledDumpTypes["cscore"] = struct{}{}
	} else {
		return nil, fmt.Errorf("unsupported product: <%s>", productName)
	}

	// Check binaries and status on Wi-Fi interfaces
	executables := make(map[string]string)
	for _, bin := range [2]string{"acs_cli", "wl"} {
		if path, err := exec.LookPath(bin); err == nil {
			executables[bin] = path
		} else {
			return nil, fmt.Errorf("missing critical binary: %s: %s", bin, err)
		}
	}
	device := &deviceInfo{
		Model: productName,
		// Interfaces:       [2]string{"wl0", "wl1"},
		EnabledDumpTypes: enabledDumpTypes,
		Executables:      executables,
	}
	if err := device.storeDeviceIDs(); err != nil {
		log.Errorf("[getCompatibleDevice] %s", err)
	}
	if err := device.getOperatingInterfaces(); err != nil {
		log.Errorf("[getCompatibleDevice] %s", err)
	}
	if err := device.getChanSpecs(); err != nil {
		log.Panicf(err.Error())
	}
	return device, nil
}

// hex->string chanspec map
var chanSpecs map[string]string
