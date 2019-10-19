package provider

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"net"
	"time"
)

// Listener for listen to event
type Listener interface {
	DoProcess(invokeID string, data string)
}

var connectionTimeout = 15 * time.Second

var (
	conn net.Conn
	out  *bufio.Writer
	in   *bufio.Reader
)

func writeShort(v int, w *bufio.Writer) error {
	var err error
	err = w.WriteByte(byte(v >> 8 & 0xFF))
	err = w.WriteByte(byte(v >> 0 & 0xFF))
	return err
}

func write(v string, w *bufio.Writer) error {
	_, err := w.Write([]byte(v))
	if err != nil {
		log.Printf("Error while sending data %v ", err)
		//TODO Implement reconnect
	}
	return err
}

func read(r *bufio.Reader, length int) ([]byte, error) {
	data := make([]byte, length)
	_, err := io.ReadFull(r, data)
	if err != nil {
		if err == io.EOF {
			log.Fatalln(err)
		}
		log.Printf("Error while reading data %v ", err)
		//TODO Implement reconnect
	}
	return data, err
}

func readShort(r *bufio.Reader) (uint16, error) {
	buff, err := read(in, 2)
	data := binary.BigEndian.Uint16(buff)
	return data, err
}

// Connect to CTI Provider
func Connect(host string, listener Listener) error {
	log.Printf("Connecting to Provider: %s...\n", host)
	var err error
	conn, err = net.DialTimeout("tcp", host, connectionTimeout)
	if err != nil {
		log.Fatalln(err)
		//TODO Implement reconnect
	}
	out = bufio.NewWriter(conn)
	in = bufio.NewReader(conn)

	responseHandler(listener)

	return err
}

// Send messaged to CTI Provider with a specific reader
func Send(invokeID string, message string) error {
	/*
	 * The Header is  8 bytes long.
	 * | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
	 * |VERSION|LENGTH |   INVOKE ID   |   XML PAYLOAD
	 */
	log.Println("=============== REQUEST ===============")
	log.Printf("(%s)\n", message)
	var err error
	err = writeShort(0, out)
	err = writeShort(len(message)+8, out)
	err = write(invokeID, out)
	err = write(message, out)
	out.Flush()
	log.Println("=======================================")
	return err
}

// ResponseHandler to handle responses
func responseHandler(listener Listener) {
	go func(listener Listener) {
		for {

			var err error
			version, err := readShort(in)
			length, err := readShort(in)
			invokeID, err := read(in, 4)
			data, err := read(in, int(length-8))

			switch err {
			case nil:
				log.Println("=============== RESPONSE ===============")
				log.Printf(" VERSION: %d\n", version)
				log.Printf("  LENGTH: %d\n", length)
				log.Printf("INVOKEID: %s\n", string(invokeID))
				log.Printf("    DATA: %s\n", string(data))
				log.Println("=======================================")
				go listener.DoProcess(string(invokeID), string(data))
			default:
				log.Fatalf("error while receiving data: %s", err)
				//TODO implement reconnect
			}

		}
	}(listener)
}

// Close the connection
func Close() {
	conn.Close()
}
