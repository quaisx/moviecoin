# moviecoin

## THIS IS WORK IN PROGRESS AT THE MOMENT

Initial version of a blockchain implementation in Go

![Moviecoin blockchain](/design/Moviecoin.png)

You will need to run a coupld of chain servers:
```
cd chainserver
go run main.go chainserver.go -p=5555
go run main.go chainserver.go -p=6666
```

and run one web wallet server:
```
cd walletserver
go run main.go walletserver.go -p=8888 -gateway=127.0.0.1 -gatewayport=5555
```

access the wallet at: `localhost:8080`

and use sender address as **"MOVIECOIN BLOCKCHAIN"**
to send yourself a few coins to get started.
