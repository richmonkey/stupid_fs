package main

import "fmt"
import "os"
import "log"
import "io"
import "net/http"
import "syscall"
import "path/filepath"

const BLOCK_SIZE = 4*1024
//const ROOT = "/Users/houxh/centos_dev/stupid_fs/data"
const ROOT = "/tmp/fs"

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

func main() {
    http.Handle("/", http.FileServer(http.Dir(ROOT)))
    http.HandleFunc("/upload/", handle_upload)

    log.Fatal(http.ListenAndServe(":8080", nil))
}

