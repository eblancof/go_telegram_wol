package wol

import (
	"encoding/hex"
	"net"
	"strings"
)

func SendWakeOnLAN(macAddress, broadcastIP string, port int) error {
	mac, err := hex.DecodeString(strings.Replace(macAddress, ":", "", -1))
	if err != nil {
		return err
	}

	var packet []byte
	packet = append(packet, []byte{255, 255, 255, 255, 255, 255}...)
	for i := 0; i < 16; i++ {
		packet = append(packet, mac...)
	}

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(broadcastIP),
		Port: port,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}
