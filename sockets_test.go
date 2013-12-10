// Unit tests for sockets
package sockets

import (
	"log"
	"testing"
	"time"
)

type ConnectionListener struct {
	monitor chan bool
	connection *TCPConnection	
}

func (connectionListener *ConnectionListener) OnData(data *[]byte) {
	log.Println("Got data")
	log.Println("Data\n",string(*data))
}

func (connectionListener *ConnectionListener) OnError(err error) {
	log.Println("on error", err)
	connectionListener.monitor <- true
}

func (connectionListener *ConnectionListener) OnDisconnect() {
	log.Println("on disconnect")	
	connectionListener.monitor <- true
}

func (connectionListener *ConnectionListener) OnConnect(connection *TCPConnection) {
	log.Println("on connect")
	connectionListener.connection = connection
}

func TestTcpConnect(t *testing.T) {	
	connectionListener := &ConnectionListener{}
	connectionListener.monitor = make(chan bool)
	TCPConnect("127.0.0.1", 1234, time.Millisecond*10, connectionListener)
	<-connectionListener.monitor
}