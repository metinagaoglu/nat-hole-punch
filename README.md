# NAT Hole Punching Experiments in Go

This repository contains a collection of Go experiments demonstrating NAT hole punching techniques. NAT hole punching is a method used to establish communication between devices behind different Network Address Translators (NATs).

## Introduction

NAT (Network Address Translation) is a common technology used in routers and firewalls to map multiple private IP addresses to a single public IP address. This can cause challenges when trying to establish direct communication between devices behind different NATs.

This project explores various NAT hole punching techniques and provides code samples and explanations for each.


### Build for osx

GOOS=darwin GOARCH=amd64 go build -o client-osx *.go

/client c

/client s

### Examples

/client c 192.168.0.11:3986 4545
./client s


### Useful resources

- https://en.wikipedia.org/wiki/Hole_punching_(networking)
- https://itnext.io/p2p-nat-traversal-how-to-punch-a-hole-9abc8ffa758e
- https://medium.com/@surapuramakhil/hole-punching-nat-in-networking-72502c8d1b7c
