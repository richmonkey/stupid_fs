#mmmmm

import socket
import struct

COMMAND_UPLOAD  = 1
COMMAND_DOWNLOAD = 2

class FS:
    def __init__(self):
        self.server_address = ('10.0.0.48', 23000)
    
    def download(self, path):
        buf = struct.pack("!iB", COMMAND_DOWNLOAD, len(path))
        buf += path
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)  
        s.connect(self.server_address)  
        s.sendall(buf)

        buf = s.recv(4)
        if len(buf) != 4:
            return ""

        size, = struct.unpack("!i", buf)

        left = size
        content = ""
        while left:
            resp = s.recv(1024*4)
            if not resp:
                break
            content += resp
            left -= len(resp)

        s.close()
        if len(content) != size:
            return ""
        else:
            return content

    def upload(self, path, data):
        buf = struct.pack("!iB", COMMAND_UPLOAD, len(path))
        buf += path

        size = len(data)
        buf += struct.pack("!i", size)
        buf += data

        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)  
        s.connect(self.server_address)  
        s.sendall(buf)
        resp = s.recv(4)
        r, = struct.unpack('!i', resp)
        return r


if __name__ == "__main__":
    import os
    fs = FS()
    p = os.path.abspath(__file__)
    f = open(p)
    data = f.read()
    #print fs.upload("/test", data)
    print fs.download("/test")
    #vvvvvvmmxxxmmm
