# MOVIECOIN

## THIS IS WORK IN PROGRESS
_(I am working on this project in my free time with the goal to build a blockchain framework to use in my other for-fun projects I am working on at the moment)_

Initial version of a blockchain implementation in Go.

![Moviecoin blockchain](/design/Moviecoin.png)

You will need to run a could of chain servers:
```
cd chainserver
go run main.go chainserver.go -p=5555
go run main.go chainserver.go -p=6666
```

and run one web wallet server that will connect to at least one mining node:
```
cd walletserver
go run main.go walletserver.go -port=8888 -gateway=127.0.0.1 -gatewayport=5555
```

access the wallet at: `localhost:8888` or whatever port value you specified for -port

and use sender address as **"MOVIECOIN BLOCKCHAIN"**
to send yourself a few coins to get started. At the moment it's an unlimited supply of coins.

What I am working on next: Web Wallet Authentication and persistence
