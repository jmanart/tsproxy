TSProxy
=======
TCP Stupid Proxy sits like a Man In The Middle of a TCP Connection between two sockets redirecting traffic both ways giving the ability to perform operations using the messages passed. Theese may not be modified, however. That may be added in a future.

Use Cases:
----------
* Redirecting spected traffic on one port to another address port
* Monitoring the traffic of a connection
* Saving communication traces

Example:
--------
tspprintlength.go demo uses LengthPrintFilter to print the direction of a message "<<<" for incoming messages and ">>>" for outgoing messages followed by every message's byte count

```
$ go run demo/tspprintlength.go 50505 127.0.0.1:25000
<<< 10
>>> 110
<<< 10
<<< 20
<<< 10
<<< 2042
```
ROADMAP:
--------
* connect/close operations
* mock connection
* extend UDP
* permit message manipulation