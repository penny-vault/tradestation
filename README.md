# pv-tradestation

pv-tradestation acts as a bridge between pvapi and the TradeStation brokerage.
It enables users to automatically execute a pvapi trading strategy within their
brokerage.

# Install

```bash
mage build
```

# Configuration

Configuration parameters should be placed in a file called `tradestation.toml`

| Field Name  | Required | Description                                                                                          |
|-------------|----------|------------------------------------------------------------------------------------------------------|
| sim         | Yes      | URL to tradestation simulated trading environment (https://sim-api.tradestation.com/v3)              |
| live        | Yes      | URL to tradestation live trading environment (https://api.tradestation.com/v3)                       |
| mode        | Yes      | Indicates whether the simulated or live api's should be used. Value should be either 'sim' or 'live' |
| state_file  | Yes      | File to save API state in (note: file is AES encrypted with SSH key)                                 |
| auth.apikey | Yes      | API Key issued by tradestation                                                                       |
| auth.secret | Yes      | API Secret issued by tradestation                                                                    |
| key_file    | No       | Path to encryption key (defaults to ~/.ssh/id_rsa)                                                   |
