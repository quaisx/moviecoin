# MOVIECOIN

## THIS IS WORK IN PROGRESS
_(I am working on this project in my free time with the goal to build a blockchain framework to use in my other for-fun projects I am working on at the moment)_

Initial version of a blockchain implementation in Go.

![Moviecoin blockchain](/design/Moviecoin.png)

You will need to run a couple of chain servers:
```
cd chainserver
go run main.go chainserver.go -p=5555
go run main.go chainserver.go -p=6666
```

and run one web wallet server that will connect to at least one mining node:
```
cd walletserver
go run main.go walletserver.go -port=8888 -gateway=127.0.0.1 -gateway_port=5555
```

access the wallet at: `localhost:8888` or whatever port value you specified for -port

and use sender address as **"MOVIECOIN BLOCKCHAIN"**
to send yourself a few coins to get started. At the moment it's an unlimited supply of coins.


![Moviecoin landing page](/design/send-1.jpeg)

![Moviecoin send from blockcahin](/design/send-2.jpeg)

![Moviecoin update wallet amount](/design/send-3.jpeg)

### Update:
Added support for AES-256 CBC encryption. Individual strings can be strongly encrypted before being saved to a database.
Encryption
![Moviecoin AES Cipher Block Chaining](/design/CBC-encryption.png)
Decryption
![Moviecoin AES Cipher Block Chaining](/design/CBC-decryption.png)

![Cipher Block Chaining Mode of Operation](https://en.wikipedia.org/wiki/Block_cipher_mode_of_operation)

![Galois/Counter Mode (GCM)](https://en.wikipedia.org/wiki/Galois/Counter_Mode)

What I am working on next: Web Wallet front with React.js
