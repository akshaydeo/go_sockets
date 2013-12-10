// This module acts as helper for socket related functions
package sockets

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// struct which wrapps underlaying connection
type TCPConnection struct {
	conn     *net.TCPConn
	isClosed bool
	isOpen   bool
}

// tcp connection listener
type TCPConnectionListener interface {
	OnConnect(channel *TCPConnection)
	OnData(data *[]byte)
	OnError(err error)
	OnDisconnect()
}

// tcp error connection
type TCPConnectionError struct {
	message string
}

// overriden method for error
func (tcpConnectionError *TCPConnectionError) Error() string {
	return tcpConnectionError.message
}

// method to send data over tcp connection
func (tcpConnection *TCPConnection) TCPSendData(data *[]byte) error {
	var err error
	_, err = tcpConnection.conn.Write(*data)
	return err
}

// checks for if error has happened and reports via corresponding tcp connection listener
func checkErrAndReport(tcpConnectionListener TCPConnectionListener, err error) bool {
	if err != nil {
		log.Panic("Error happened", err)
		if tcpConnectionListener != nil {
			tcpConnectionListener.OnError(err)
		}
		return true
	}
	return false
}

// method to close the connection
func (tcpConnection *TCPConnection) CloseConnection() error {	
	// check if connection is already closed
	if !tcpConnection.isOpen && tcpConnection.isClosed{
		return nil
	}
	tcpConnection.isOpen = false
	tcpConnection.isClosed = true
	return tcpConnection.conn.Close()	
}

// method to check if connection is open
func (tcpConnection *TCPConnection) IsOpen() bool {
	return tcpConnection.isOpen
}

// method to make tcp connection
func TCPConnect(host string, port int, timeout time.Duration, tcpConnectionListener TCPConnectionListener) {
	address := fmt.Sprintf("%s:%d", host, port)
	log.Println("Address for tcp connection is ", address)	
	tcpAddress, err := net.ResolveTCPAddr("tcp", address)
	if checkErrAndReport(tcpConnectionListener, err) {		
		return
	}
	tcpConnection := TCPConnection{}
	tcpConnection.conn, err = net.DialTCP("tcp", nil, tcpAddress)
	log.Println("Connection made successfully")
	tcpConnection.isClosed = false
	tcpConnection.isOpen = true
	if checkErrAndReport(tcpConnectionListener, err) {
		tcpConnection.CloseConnection()
		return
	}
	// set connection keeplive
	err = tcpConnection.conn.SetKeepAlive(true)
	if checkErrAndReport(tcpConnectionListener, err) {
		tcpConnection.CloseConnection()
		return
	}
	// set connection nodelay
	err = tcpConnection.conn.SetNoDelay(true)
	if checkErrAndReport(tcpConnectionListener, err) {
		tcpConnection.CloseConnection()
		return
	}
	tcpConnectionListener.OnConnect(&tcpConnection)
	// start routine for readin data and triggering callback if callback is passed
	if tcpConnectionListener != nil {
		go func() {
			for {
				if !tcpConnection.isOpen {
					break
				}
				log.Println("Reading on TCP connection")
				buf := make([]byte, 2048)
				_, err = tcpConnection.conn.Read(buf)
				if err != nil {
					if err == io.EOF {
						tcpConnectionListener.OnDisconnect()
					} else {
						tcpConnectionListener.OnError(err)
					}
					break
				}
				tcpConnectionListener.OnData(&buf)
			}
			tcpConnection.CloseConnection()
		}()
	}
}