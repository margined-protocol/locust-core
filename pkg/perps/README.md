# Perps

Perps is a package that tries to abstract the interfaces of Perp DEXes, rn only in Cosmos.

## Useful Info && Queries

### dYdX

```bash
curl -s -X GET https://indexer.dydx.trade/v4/addresses/dydx1ha2hjlce7sqp59g8xhxz2jds97x8fdw9mrf4j3/subaccountNumber/0 | jq .
curl -s -X GET "https://indexer.dydx.trade/v4/fills?address=dydx1ha2hjlce7sqp59g8xhxz2jds97x8fdw9mrf4j3&subaccountNumber=0" | jq .
curl -s -X GET https://indexer.dydx.trade/v4/orderbooks/perpetualMarket/ATOM-USD | jq .
curl -s -X GET https://indexer.dydx.trade/v4/candles/perpetualMarkets/ATOM-USD?resolution=1MIN | jq .
```
