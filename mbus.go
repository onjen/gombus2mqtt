package main

import (
	"fmt"
	"time"

	"github.com/jonaz/gombus"
)

func fetchValue(device string, address int) (*gombus.DecodedFrame, error) {
	conn, err := gombus.DialSerial(device)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial serial to device %v: %v", device, err)

	}
	_, err = conn.Write(gombus.SndNKE(uint8(address)))
	if err != nil {
		return nil, fmt.Errorf("Failed to write NKE: %v", err)
	}
	err = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		return nil, fmt.Errorf("Failed to set read deadline: %v", err)
	}
	_, err = gombus.ReadSingleCharFrame(conn)
	if err != nil {
		return nil, fmt.Errorf("Failed to set read single char frame: %v", err)
	}
	frame, err := gombus.ReadSingleFrame(conn, address)
	if err != nil {
		return nil, fmt.Errorf("Failed to set read single frame: %v", err)
	}
	return frame, nil
}
