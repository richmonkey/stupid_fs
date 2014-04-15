package main
//xxx
import "math/rand"

//import "hash/crc32"
//import "encoding/base64"
//import "bytes"
//import "syscall"
//import "strconv"

import "encoding/binary"

import "fmt"
import "os"
import "net"
import "io"
import "time"
import "log"

const BLOCK_SIZE = 4*1024
const ROOT = "/vagrant/stupid_fs/data/"

func handle_upload(conn *net.TCPConn) bool {
    var len uint8
    var path string
    var file *os.File
    var size int32
    var n int64

	buff := make([]byte, 1)
	_, err := io.ReadFull(conn, buff)
	if err != nil {
        goto Error
	}

    len = buff[0]
    if len == 0 {
        goto Error
    }

    buff = make([]byte, len)
    _, err = io.ReadFull(conn, buff)
    if err != nil {
        goto Error
    }
    path = ROOT + string(buff)


    file, err = os.Create(path)
    if err != nil {
        fmt.Println("create path error:", err)
        goto Error
    }
    defer file.Close()

    err = binary.Read(conn, binary.BigEndian, &size)
    if err != nil {
        goto Error
    }

    fmt.Printf("upload path:%s size:%d\n", path, size)
    
    n, err = io.CopyN(file, conn, int64(size))
    if err != nil || n != int64(size) {
        goto Error
    }

    file.Sync()
    binary.Write(conn, binary.BigEndian, int32(0))
    return true

Error:
    binary.Write(conn, binary.BigEndian, int32(-1))
    return false
}

type writerOnly struct {
	io.Writer
}

func handle_download(conn *net.TCPConn) bool {
    var path string
    var len uint8
    var f *os.File
    var fi os.FileInfo
    var size int32
    var n int64

	buff := make([]byte, 1)
	_, err := io.ReadFull(conn, buff)
	if err != nil {
        goto Error
	}
    len = buff[0]
    if len == 0 {
        goto Error
    }
    buff = make([]byte, len)
    _, err = io.ReadFull(conn, buff)
    if err != nil {
        goto Error
    }
    path = ROOT + string(buff)
    f, err = os.Open(path)
    if err != nil {
        goto Error
    }
    defer f.Close()
    fi, err = f.Stat()
    if err != nil {
        goto Error
    }
    size = int32(fi.Size())
    fmt.Printf("download path:%s size:%d\n", path, size)
    err = binary.Write(conn, binary.BigEndian, size)
    if err != nil {
        return false
    }
    n, err = io.Copy(writeOnly{conn}, f)
    if err != nil {
        return false
    }
    if n != int64(size) {
        return false
    }

    return true
Error:
    binary.Write(conn, binary.BigEndian, int32(-1))
    return false
}


const COMMAND_UPLOAD  = 1
const COMMAND_DOWNLOAD = 2
func handle_client(conn *net.TCPConn) bool {
    defer conn.Close()

    var cmd int32
    err := binary.Read(conn, binary.BigEndian, &cmd)
    if err != nil {
        return false
    }
    if cmd == COMMAND_UPLOAD {
        return handle_upload(conn)
    } else if cmd == COMMAND_DOWNLOAD {
        return handle_download(conn)
    } else {
        log.Println("unknown cmd:", cmd)
        return false
    }
}

func main() {
    log.SetFlags(log.Lshortfile|log.LstdFlags)
    rand.Seed(time.Now().UnixNano())
    ip := net.ParseIP("0.0.0.0")
    addr := net.TCPAddr{ip, 23000, ""}

    listen, err := net.ListenTCP("tcp", &addr);
    if err != nil {
        fmt.Println("初始化失败", err.Error())
        return
    }
    for {
        client, err := listen.AcceptTCP();
        if err != nil {
            return
        }
        go handle_client(client)
    }

}
