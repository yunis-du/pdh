# PDH

send things from one computer to another

## install

### Download for your system
```bash
https://github.com/duyunzhi/pdh/releases
```
Or gitee
```bash
https://gitee.com/duyunzhi_admin/pdh/releases
```

### On macOS you can install the latest release with Homebrew
```bash
brew tap duyunzhi/brew
brew install pdh
```

## Usage

### simple send and receive
send a file

```bash
pdh send [files or folder]

...
share code is: xxxx-xxxx-xxxx-xxxx
...
```

receive
```bash
pdh receive xxxx-xxxx-xxxx-xxxx
```

### deployment your owner relay

```bash
docker pull duyunzhi1/pdh-relay

docker run -it -d -p 6880:6880 --name=pdh-relay --restart=always duyunzhi1/pdh-relay
```

send
```bash
pdh send --relay 'your relay' [files or folder]
```

receive
```bash
pdh receive --relay 'your relay' xxxx-xxxx-xxxx-xxxx
```

## License
MIT

## Statement
This project refers to [croc](https://github.com/schollz/croc), used the code related to file processing, etc.