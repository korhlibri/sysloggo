import socket

HOST = "localhost"
UDPPORT = 514
TCPPORT = 6514

invalid_log = "Hello!"
valid_log = "<34>1 2024-01-01T00:00:00.162Z pythontest test - ID01 - Hello World!"

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock.sendto(bytes(invalid_log, "utf-8"), (HOST, UDPPORT))
sock.sendto(bytes(valid_log, "utf-8"), (HOST, UDPPORT))

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect((HOST, TCPPORT))
sock.sendall(bytes(invalid_log, "utf-8"))
sock.close()

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect((HOST, TCPPORT))
sock.sendall(bytes(valid_log, "utf-8"))
sock.close()