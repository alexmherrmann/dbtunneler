package tun

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

func bufferingCancelableCopy(
	ctx context.Context,
	dst io.Writer,
	src io.Reader,
) (written int64, copyErr error) {
	buf := make([]byte, 64*1024)
	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
			nr, readErr := src.Read(buf)
			if nr > 0 {
				nw, writeErr := dst.Write(buf[0:nr])
				if nw > 0 {
					written += int64(nw)
				}
				if writeErr != nil {
					copyErr = writeErr
					return
				}
				if nr != nw {
					copyErr = io.ErrShortWrite
					return
				}
			}
			if readErr != nil {
				if readErr != io.EOF {
					copyErr = readErr
				}
				return
			}
		}
	}
}

// Handle forwarding connections to and from an ssh tunnel
func handleSshTunnelConnections(
	ctx context.Context,
	localConn net.Conn,
	remoteConn net.Conn,
	// A channel to send errors to
	errChan chan<- error,
) {

	localToRemoteDone := make(chan bool)
	remoteToLocalDone := make(chan bool)

	// Forward localConn to remoteConn
	go func() {
		defer close(errChan)
		_, err := bufferingCancelableCopy(ctx, remoteConn, localConn)
		if err != nil {
			errChan <- err
		}
		localToRemoteDone <- true
	}()

	// Forward remoteConn to localConn
	go func() {
		defer close(errChan)
		_, err := bufferingCancelableCopy(ctx, localConn, remoteConn)
		if err != nil {
			errChan <- err
		}
		remoteToLocalDone <- true
	}()

	<-localToRemoteDone
	<-remoteToLocalDone

}
func StartSSHTunnel(
	ctx context.Context,
	config *ssh.ClientConfig,
	sshHost string,
	localPort string,
	remoteHost string,
	remotePort string,
) (eventualErr <-chan error, startErr error) {

	errChan := make(chan error)

	eventualErr = errChan

	sshConn, err := ssh.Dial("tcp", sshHost, config)
	if err != nil {
		return nil, err
	}

	// Listen on local port
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", localPort))
	if err != nil {
		return nil, err
	}

	// Handle incoming connections on reverse forwarded tunnel
	go func() {
		defer listener.Close()
		defer sshConn.Close()

		go func() {
			<-ctx.Done()
			listener.Close()
		}()

		for {

			localConn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %s", err)
				break
			}

			forwardConn, err := sshConn.Dial("tcp", fmt.Sprintf("%s:%s", remoteHost, remotePort))

			if err != nil {
				log.Printf("Error dialing remote host: %s", err)
				continue
			}

			go handleSshTunnelConnections(ctx, forwardConn, localConn, errChan)

		}
	}()

	return nil, nil
}
