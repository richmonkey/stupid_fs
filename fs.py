#mmmmm

import socket
import struct
import sys
import requests
import time

class config:
    FS_PATH = "http://172.25.1.111:8080"

class FS:
    @classmethod
    def upload(self, path, data):
        resp = requests.post(config.FS_PATH + '/upload' + path, data)
        return resp.status_code == 200

    @classmethod
    def download(self, path):
        resp = requests.get(config.FS_PATH + path)
        return resp.content if resp.status_code == 200 else ''

    @classmethod
    def upload_range(self, path, data, r):
        headers = {"Range":"bytes=%d-%d"%(r[0], r[1])}
        resp = requests.post(config.FS_PATH + '/range_upload' + path, data, headers=headers)
        return resp.status_code == 200

    @classmethod
    def rename(self, src, dst):
        obj = {"src":src, "dst":dst}
        url = config.FS_PATH + "/rename"
        resp = requests.post(url, data=obj)
        return resp.status_code == 200

    @classmethod
    def remove(self, path):
        resp = requests.delete(config.FS_PATH + '/remove' + path)
        return resp.status_code == 200

    @classmethod
    def exists(self, path):
        url = config.FS_PATH + path        
        resp = requests.head(url)
        return (resp.status_code == 200)
        
if __name__ == "__main__":
    import os
    config.FS_PATH = "http://192.168.33.10:8083"
    fs = FS()
    fs.upload_range("/test_range", "1111", (8, 11))
    fs.upload_range("/test_range", "11111111", (0, 7))

    p = os.path.abspath(__file__)
    f = open("/tmp/test", "wb")
    
    FILE_SIZE = 600*1024
    f.seek(FILE_SIZE - 1)
    f.write("1")
    f.close()
    f = open("/tmp/test", "rb")
    data = f.read()
    print len(data)
    b = time.time()
    fs.upload("/test", data)
    e = time.time()
    print "time:", e-b
    print len(fs.download("/test"))
    print "exists:", fs.exists("/test")
    print "exists:", fs.exists("/test2")    
    print fs.rename("/test", "/test2")
    print fs.remove("/test2")


    #vvvvvvmmxxxmmm


