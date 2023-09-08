package main

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// List of XML RPC getter function codes.
const (
	FnLogin  = "15"
	FnLogout = "16"
)

// List of XML RPC setter function codes.
const (
	FnCMSystemInfo = "2"
	FnCMState      = "136"
)

// List of string constants from the XML API responses.
const (
	OperStateOK          = "OPERATIONAL"
	NetworkAccessAllowed = "NetworkAccess"
)

// CMSystemInfo shows cable modem system info.
type CMSystemInfo struct {
	DocsisMode      string `xml:"cm_docsis_mode"`
	HardwareVersion string `xml:"cm_hardware_version"`
	MacAddr         string `xml:"cm_mac_addr"`
	SerialNumber    string `xml:"cm_serial_number"`
	SystemUptime    int    `xml:"cm_system_uptime"`
	NetworkAccess   string `xml:"cm_network_access"`
}

// UnmarshalXML is a standard unmarshaller + string to seconds convertor.
func (c *CMSystemInfo) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias CMSystemInfo
	aux := &struct {
		*Alias
		SystemUptime string `xml:"cm_system_uptime"`
	}{
		Alias: (*Alias)(c),
	}

	if err := d.DecodeElement(&aux, &start); err != nil {
		return err //nolint:wrapcheck
	}

	dur, err := parseDuration(aux.SystemUptime)
	if err != nil {
		return err
	}
	c.SystemUptime = int(dur.Seconds())

	return nil
}

// CMState shows cable modem state.
type CMState struct {
	TunnerTemperature int      `xml:"TunnerTemperature"`
	Temperature       int      `xml:"Temperature"`
	OperState         string   `xml:"OperState"`
	WANIPv4Addr       string   `xml:"wan_ipv4_addr"`
	WANIPv6Addrs      []string `xml:"wan_ipv6_addr>wan_ipv6_addr_entry"`
}

// UnmarshalXML is a standard unmarshaller + fahrenheit to celsius convertor.
func (c *CMState) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias CMState
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := d.DecodeElement(&aux, &start); err != nil {
		return err //nolint:wrapcheck
	}
	c.TunnerTemperature = fahrenheitToCelsius(c.TunnerTemperature)
	c.Temperature = fahrenheitToCelsius(c.Temperature)

	return nil
}

var durationRegexp = regexp.MustCompile(`(?:(\d+)day\(s\))?(\d+)h:(\d+)m:(\d+)s`)

// Input format: "1day(s)2h:34m:56s".
func parseDuration(s string) (time.Duration, error) {
	matches := durationRegexp.FindStringSubmatch(s)
	if len(matches) != 5 {
		return 0, fmt.Errorf("invalid duration string")
	}

	days, _ := strconv.Atoi(matches[1])
	hours, _ := strconv.Atoi(matches[2])
	minutes, _ := strconv.Atoi(matches[3])
	seconds, _ := strconv.Atoi(matches[4])

	dur := time.Duration(days)*24*time.Hour +
		time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second

	return dur, nil
}

func fahrenheitToCelsius(f int) int {
	return (f - 32) * 5.0 / 9
}
