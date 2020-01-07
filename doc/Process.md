# Adrestia Process Documentation

![Process](img/adrestia_processor.png)

## handleCreateOrder()
1. Broadcasts transaction that sends from HotWallet to exchange.
2. Saves txid.

## handleNthExchange()
1. Checks sent amount has arrived to exchange an it is usable.
2. If balance has arrived, creates selling order at market price.

## handleConversion()
1. Checks the order has been completed.
2. Checks if it is the last conversion or requires a secon one. 
    a. if it is last conversion marls order as ExchangeComplete
    b. Withdraws converted BTC to second exchange.
    
## handleCompletedExchange()
1. Sends converted assets to target HotWallet.
