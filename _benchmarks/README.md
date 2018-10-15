# Benchmarks

Last updated on October 15, 2018.

## Hardware

* Processor: Intel(R) Core(TM) **i7-8750H** CPU @ 2.20GHz 2.20GHz
* RAM: **16.00 GB**

## Software

* OS: Microsoft **Windows 10** [Version 1803 (OS Build 17134.345)]
* HTTP Benchmark Tool: https://github.com/codesenberg/bombardier, latest version **1.2.0**
* Go Version: 1.11.1
* **muxie**: https://github.com/kataras/muxie, latest version **1.0.0**
    * Bench code: [muxie/main.go](muxie/main.go)
* **httprouter**: https://github.com/julienschmidt/httprouter, latest version **1.2.0**
    * Bench code: [httprouter/main.go](httprouter/main.go)
* **gin**: https://github.com/gin-gonic/gin, latest version **1.3.0**
    * Bench code: [gin/main.go](gin/main.go)
* **gorilla mux**: https://github.com/gorilla/mux, latest version **1.6.2**
    * Bench code: [gorilla-mux/main.go](gorilla-mux/main.go)

## Results

### Static Path

```sh
bombardier -c 125 -n 1000000 http://localhost:3000
```

#### Muxie

![](static_path_muxie.png)

#### Httprouter

![](static_path_httprouter.png)

#### Gin

![](static_path_gin.png)

#### Gorilla Mux

![](static_path_gorilla-mux.png)

### Parameterized (dynamic) Path

```sh
bombardier -c 125 -n 1000000 http://localhost:3000/user/42
```

#### Muxie

![](parameterized_path_muxie.png)

#### Httprouter

![](parameterized_path_httprouter.png)

#### Gin

![](parameterized_path_gin.png)

#### Gorilla Mux

![](parameterized_path_gorilla-mux.png)