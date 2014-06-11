package main

import "fmt"
import "os"
import "log"
import "io"
import "net/http"
import "syscall"
import "path/filepath"
import "strconv"
import "github.com/jimlawless/cfg"

const BLOCK_SIZE = 4*1024
var ROOT = "/tmp/fs"
var PORT = 8080
var BIND_ADDR = ""
func create_file(path string) (*os.File, error) {
    file, err := os.Create(path)
    if err != nil {
        e, _ := err.(*os.PathError)
        errno, _ :=  e.Err.(syscall.Errno)
        if errno == syscall.ENOTDIR || errno == syscall.ENOENT {
            dir, _ := filepath.Split(path)
            err = os.MkdirAll(dir, 0777)
            if err == nil {
                file, err = os.Create(path)
                return file, err
            } else {
                return nil, err
            }
        } else {
            return nil, err
        }
    }
    return file, nil
}

func handle_upload(w http.ResponseWriter, r *http.Request) {
    var n int64
    len := r.ContentLength
    body := r.Body

    //7==len("/upload")
    path := ROOT + r.URL.Path[7:]
    
    file, err := create_file(path)
    if err != nil {
        fmt.Println("create path error:", err)
        goto Error
    }
    defer file.Close()
    
    fmt.Printf("upload path:%s\n", path)
    
    n, err = io.CopyN(file, body, len)
    if err != nil || n != len {
        goto Error
    }
    w.WriteHeader(http.StatusOK)
    return
Error:
    w.WriteHeader(400)
}

func read_cfg() {
    app_cfg := make(map[string]string)
	err := cfg.Load("fs.cfg", app_cfg)
	if err != nil {
		log.Fatal(err)
	}
    root, present := app_cfg["root"]
    if !present {
        fmt.Println("need config root directory")
        os.Exit(1)
    }
    ROOT = root

    port, present := app_cfg["port"]
    if !present {
        fmt.Println("need config listen port")
        os.Exit(1)
    }
    nport, err := strconv.Atoi(port)
    if err != nil {
        fmt.Println("need config listen port")
        os.Exit(1)
    }
    PORT = nport
    if _, present = app_cfg["bind_addr"]; present {
        BIND_ADDR = app_cfg["bind_addr"]
    }
	fmt.Printf("root:%s bind addr:%s port:%d\n", ROOT, BIND_ADDR, PORT)
}

func main() {
    read_cfg()
    http.Handle("/", http.FileServer(http.Dir(ROOT)))
    http.HandleFunc("/upload/", handle_upload)
    addr := fmt.Sprintf("%s:%d", BIND_ADDR, PORT)
    log.Fatal(http.ListenAndServe(addr, nil))
}

