package main

import (
	"errors"
	"fmt"
	"go-tunnel-proxy/cache"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

func handleConn(clientConn net.Conn, cache *cache.LRUCache) {
	defer clientConn.Close()
	buf := make([]byte, 4096)
	_, err := clientConn.Read(buf)
	if err != nil {
		return
	}
	data := string(buf)
	if !strings.Contains(data, "CONNECT") {
		return
	}
	clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	url := strings.Split(data, " ")[1]
	var mu sync.Mutex
	mu.Lock()
	resp := cache.Get(url)
	mu.Unlock()
	if resp != nil {
		fmt.Println("fetched from cache", url)
		clientConn.Write(resp)
		return
	}

	targetConn, err := net.Dial("tcp", url)
	if err != nil {
		log.Panic(err)
		return
	}
	defer targetConn.Close()

	fmt.Println("request to", url)

	var wg sync.WaitGroup
	wg.Add(2)
	go tunnelConn(targetConn, clientConn, &wg)
	go func() {
		defer wg.Done()
		copyBuffer(clientConn, targetConn, cache, url, &mu)
	}()
	wg.Wait()

}

var ErrShortWrite = errors.New("short write")
var errInvalidWrite = errors.New("invalid write result")
var ErrShortBuffer = errors.New("short buffer")
var EOF = errors.New("EOF")

func copyBuffer(dst, src net.Conn, cache *cache.LRUCache, url string, mu *sync.Mutex) (written int64, err error) {
	size := 32 * 1024
	buf := make([]byte, size)
	total := 0
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			// mu.Lock()
			// cache.Put(url, buf)
			// mu.Unlock()
			// fmt.Println("added to cache", url)
			total += nr
			// fmt.Println(url, "size", total, size)
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != EOF {
				err = er
			}
			break
		}
	}
	fmt.Println(url, "completed size", total, size)
	return written, err
}
func tunnelConn(dst, src net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	io.Copy(dst, src)
}
func main() {
	cache := cache.NewLRUCache(10)
	listener, _ := net.Listen("tcp", ":8080")
	defer listener.Close()
	for {
		conn, _ := listener.Accept()
		go handleConn(conn, cache)
	}

}
