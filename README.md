# MAYO in Go
*Andreas Skriver Nielsen, Markus Valdemar Grønkjær Jensen, and Hans-Christian Kjeldsen*

## Overview
[MAYO](https://pqmayo.org/assets/specs/mayo-round2.pdf) is a digital signature scheme based on the Oil and Vinegar (O&V) multivariate 
quadratic framework, originally introduced by Beullens in 2022. It improves upon O&V by significantly reducing public key 
sizes while maintaining security. The key features of MAYO includes:
- Based on Oil and Vinegar: Uses multivariate quadratic polynomials for signature generation.
- Smaller public keys: Reduces key size by modifying the trapdoor structure and introducing a whipping technique.
- Whipped map construction: Expands the original O&V signature scheme by mapping it into a higher-dimensional space using structured public matrices.
- Efficient signing and verification: Despite smaller public keys, the signature generation process remains efficient by ensuring sufficient solvability conditions.

MAYO offers a trade-off between public key size and signature size, making it a promising candidate for post-quantum cryptographic applications.

## Installation
As a prerequisite make sure to have installed Go, which can be downloaded [here](https://go.dev/doc/install). Afterwards download the the project as a ZIP, or clone the repository from source:
```
$ git clone https://github.com/AU-HC/mayo-go
```
Then get the dependencies:
```
$ cd mayo-go
$ go ge
```
Lastly, our implementation utilizes functions written in C, thus a C compiler is needed, for Windows we recommend: 
[tdm-gcc](https://jmeubank.github.io/tdm-gcc/articles/2021-05/10.3.0-release). Afterwards, remember to enable CGO:
```
go env -w CGO_ENABLED=1
```

## Usage
Our implementation is currently a command line tool, to generate keys, sign, and verify a message the following command can be executed:
```
$ go run main.go -p=2
```
It's important to note that the `-p` flag must be set, as it specifies the parameter set of MAYO.

Our implementation also has alternate options which can be set, using the following flags:
- `-b` of type `int`: Specifies the amount of samples for a benchmarking run. Setting this flag with a value other than 0.

## Remarks
- This branch has some optimized code, which uses bit-sliced arithmetic on slices.
- See [optimized-implementation-arrays](https://github.com/AU-HC/mayo-go/tree/optimized-implementation-arrays) for an optimized implementation that uses bit-sliced arithmetic on arrays.
- See [master](https://github.com/AU-HC/mayo-go/tree/master) for unoptimized code, which is based heavily the specification.
