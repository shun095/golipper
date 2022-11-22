# golipper
Simple clipboard server for ssh

## Setup ssh config like following

```
Host server_name
    HostName XXX.XXX.XXX.XXX
    Port XXXX
    User XXXX
    IdentityFile ~/.ssh/xxx_id_rsa
    ...
    # Add following line
    RemoteForward 8377 localhost:8377
```

## Launch golipper at local

```sh
# Local
go run src/main.go
```

## Send to clipboard remotely

```sh
# Remote
echo "hello world" | nc localhost 8377
```
