package main

import "encoding/xml"

// CMState shows cable modem state.
type CMState struct {
	TunnerTemperature int      `xml:"TunnerTemperature"`
	Temperature       int      `xml:"Temperature"`
	OperState         string   `xml:"OperState"`
	WANIPv4Addr       string   `xml:"wan_ipv4_addr"`
	WANIPv6Addr       []string `xml:"wan_ipv6_addr>wan_ipv6_addr_entry"`
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

func fahrenheitToCelsius(f int) int {
	return (f - 32) * 5.0 / 9
}
