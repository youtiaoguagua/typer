<div align="center">
<img width="200" src="https://user-images.githubusercontent.com/30404329/217255681-4434dddd-939c-4330-8c7e-1d3f1013e6d0.gif">
</div>

# Typer

Practice typing in the terminal with ssh and play against with others.

## Play

play with ssh

```shell
ssh [<name>@]<host> -p <port>
```

use the example server
```shell
ssh change-your-user@typer.gaobili.cn -p 7788
```

### offline mod

* choose whether to include number
* select input length
* select restart

![offline](https://user-images.githubusercontent.com/30404329/217268831-c028e083-179a-4e81-92f6-b50cc50dba71.gif)

### online mod

* click online mod to play with others

![online](https://user-images.githubusercontent.com/30404329/217282612-7716ea91-8d84-41e2-9f20-431208c32631.gif)


## Install

### docker

```shell
docker run -it typer registry.cn-hangzhou.aliyuncs.com/youtiaoguagua/typer:latest
```

### docker-compose

```yml
version: '3.6'
services:
  typer:
    image: registry.cn-hangzhou.aliyuncs.com/youtiaoguagua/typer:latest
    restart: always
    ports:
      - "7788:7788"
    volumes:
      - ./data:/typer/data
    tty: true
```

### source

```shell
go install github.com/youtiaoguagua/typer@latest

typer
```
