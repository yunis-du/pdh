# PDH

send things from one computer to another

![image](pdh_demo.gif)

## install

### Download for your system
```bash
https://github.com/duyunis/pdh/releases
```
Or gitee
```bash
https://gitee.com/duyunis_admin/pdh/releases
```

### On macOS you can install the latest release with Homebrew
```bash
brew tap duyunis/brew
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
docker pull duyunis/pdh-relay:latest

docker run -it -d -p 6880:6880 --name=pdh-relay --restart=always duyunis/pdh-relay:latest
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