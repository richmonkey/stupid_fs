package main

import "fmt"
import "os"
import "io"
import "net/http"
import "syscall"
import "path/filepath"
import "strconv"
import "github.com/richmonkey/cfg"
import "strings"
import "runtime"
import "flag"
import log "github.com/golang/glog"


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

func open_file(path string) (*os.File, error) {
    file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        e, _ := err.(*os.PathError)
        errno, _ :=  e.Err.(syscall.Errno)
        if errno == syscall.ENOTDIR || errno == syscall.ENOENT {
            dir, _ := filepath.Split(path)
            err = os.MkdirAll(dir, 0777)
            if err == nil {
                file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
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
        log.Info("create path error:", err)
        goto Error
    }
    defer file.Close()
    
    log.Infof("upload path:%s\n", path)
    
    n, err = io.CopyN(file, body, len)
    if err != nil || n != len {
        goto Error
    }
    w.WriteHeader(http.StatusOK)
    return
Error:
    w.WriteHeader(400)
}

func parse_range(header http.Header) (int64, int64){
    region := header.Get("Range")
    if region == "" {
        log.Info("no range header")
        return -1, -1
    }
    index := strings.Index(region, "bytes=")
    if index == -1 {
        return -1, -1
    }
    r := region[index+6:]
    ary := strings.Split(r, "-")
    if len(ary) != 2 {
        return -1, -1
    }
    begin, err := strconv.ParseInt(ary[0], 10, 64)
    if err != nil {
        return -1, -1
    }
    end, err := strconv.ParseInt(ary[1], 10, 64)
    if err != nil {
        return -1, -1
    }
    return begin, end
}

func handle_range_upload(w http.ResponseWriter, r *http.Request) {
    var n int64
    var begin, end int64

    len := r.ContentLength
    body := r.Body

    //13==len("/range_upload")
    path := ROOT + r.URL.Path[13:]

    file, err := open_file(path)
    if err != nil {
        log.Info("create path error:", err)
        goto Error
    }
    defer file.Close()
    log.Info("upload path:%s\n", path)

    begin, end = parse_range(r.Header)
    if begin == -1 && end == -1 {
        log.Info("no range header")
        goto Error
    }
    if (end - begin + 1 != len) {
        log.Info("content length not equal range")
        goto Error
    }

    _, err = file.Seek(begin, os.SEEK_SET)
    if err != nil {
        log.Info("seek error", err)
        goto Error
    }

    n, err = io.CopyN(file, body, len)
    if err != nil || n != len {
        goto Error
    }
    w.WriteHeader(http.StatusOK)
    return
Error:
    w.WriteHeader(400)
}

func handle_mv(w http.ResponseWriter, r *http.Request) {
	var err error
	src := r.FormValue("src")
	dst := r.FormValue("dst")
	if src == "" || dst == "" {
		log.Infof("rename err:%s %s", src, dst)
		goto Error
	}
	
	src = ROOT + src
	dst = ROOT + dst
	log.Infof("rename:%s %s", src, dst)
	
	err = os.Rename(src, dst)
	if err != nil {
		log.Infof("rename err:", err)
		goto Error
	}
    w.WriteHeader(http.StatusOK)
    return
Error:
    w.WriteHeader(400)

}

func handle_del(w http.ResponseWriter, r *http.Request) {
    //7==len("/remove")
    path := ROOT + r.URL.Path[7:]
	err := os.Remove(path)
	if err != nil {
		log.Info("remove error:", err)
		goto Error
	}
	log.Info("remove:", path)
    w.WriteHeader(http.StatusOK)
    return
Error:
    w.WriteHeader(400)
}

func read_cfg(path string) {
    app_cfg := make(map[string]string)
	err := cfg.Load(path, app_cfg)
	if err != nil {
		log.Fatal(err)
	}
    root, present := app_cfg["root"]
    if !present {
        log.Info("need config root directory")
        os.Exit(1)
    }
    ROOT = root

    port, present := app_cfg["port"]
    if !present {
        log.Info("need config listen port")
        os.Exit(1)
    }
    nport, err := strconv.Atoi(port)
    if err != nil {
        log.Info("need config listen port")
        os.Exit(1)
    }
    PORT = nport
    if _, present = app_cfg["bind_addr"]; present {
        BIND_ADDR = app_cfg["bind_addr"]
    }
	log.Info("root:%s bind addr:%s port:%d\n", ROOT, BIND_ADDR, PORT)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Info("usage:http_fs config")
		return
	}

    read_cfg(flag.Args()[0])

    http.Handle("/", http.FileServer(http.Dir(ROOT)))
    http.HandleFunc("/upload/", handle_upload)
    http.HandleFunc("/range_upload/", handle_range_upload)
	http.HandleFunc("/remove/", handle_del)
	http.HandleFunc("/rename", handle_mv)

    addr := fmt.Sprintf("%s:%d", BIND_ADDR, PORT)
    log.Fatal(http.ListenAndServe(addr, nil))
}

