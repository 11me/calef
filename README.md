# Calef - crypto alef
Monitors crypto assets on binance.

## Quick start
```console
$ make up
$ make run
```

Subscribe to NATS
```console
$ nats sub binancef.ticks.btcusdt
$ nats sub binancef.ticks.ethusdt

$ nats sub binancef.bars.1m.btcusdt
$ nats sub binancef.bars.1m.ethusdt
```
