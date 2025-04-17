# Nolus Yield Market

This document holds some information about the Nolus Yield Market.

## Contracts

```json
{
  "USDC_NOBLE": "nolus1ueytzwqyadm6r0z8ajse7g6gzum4w3vv04qazctf8ugqrrej6n4sq027cf",
  "USDC": "nolus1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3sqaa3c5"
}
```

## Useful Commands

```bash
CONTRACT_ADDRESS=
NODE=https://nolus-rpc.publicnode.com:443
```

```bash
nolusd q wasm contract-state smart nolus1ueytzwqyadm6r0z8ajse7g6gzum4w3vv04qazctf8ugqrrej6n4sq027cf "{\"config\": []}" --node $NODE --output json
nolusd q wasm contract-state smart nolus1ueytzwqyadm6r0z8ajse7g6gzum4w3vv04qazctf8ugqrrej6n4sq027cf "{\"deposit_capacity\": []}" --node $NODE --output json
nolusd q wasm contract-state smart nolus1ueytzwqyadm6r0z8ajse7g6gzum4w3vv04qazctf8ugqrrej6n4sq027cf "{\"lpp_balance\": []}" --node $NODE --output json
nolusd q wasm contract-state smart nolus1ueytzwqyadm6r0z8ajse7g6gzum4w3vv04qazctf8ugqrrej6n4sq027cf "{\"config\": []}" --node $NODE --output json
nolusd q wasm contract-state smart nolus1ueytzwqyadm6r0z8ajse7g6gzum4w3vv04qazctf8ugqrrej6n4sq027cf "{\"price\": []}" --node $NODE --output json
nolusd q wasm contract-state smart nolus1ueytzwqyadm6r0z8ajse7g6gzum4w3vv04qazctf8ugqrrej6n4sq027cf "{\"balance\": {\"address\":\"nolus1tarzv9ehnhawqz0h8qc3gzm6v3zfm4vm3eh33a\"}}" --node $NODE --output json
nolusd q wasm contract-state smart nolus1ueytzwqyadm6r0z8ajse7g6gzum4w3vv04qazctf8ugqrrej6n4sq027cf "{\"rewards\": {\"address\":\"nolus1tarzv9ehnhawqz0h8qc3gzm6v3zfm4vm3eh33a\"}}" --node $NODE --output json
nolusd q wasm contract-state smart nolus1ueytzwqyadm6r0z8ajse7g6gzum4w3vv04qazctf8ugqrrej6n4sq027cf "{\"rewards\": {\"address\":\"nolus1tarzv9ehnhawqz0h8qc3gzm6v3zfm4vm3eh33a\"}}" --node $NODE --output json
nolusd q wasm contract-state smart nolus1ueytzwqyadm6r0z8ajse7g6gzum4w3vv04qazctf8ugqrrej6n4sq027cf "{\"quote\": {\"amount\":\"10000\"}}" --node $NODE --output json | jq .
```
