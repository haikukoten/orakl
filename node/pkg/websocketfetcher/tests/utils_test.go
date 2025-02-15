package tests

import (
	"encoding/json"
	"strings"
	"testing"

	"bisonai.com/miko/node/pkg/websocketfetcher/common"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/binance"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/bitget"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/bithumb"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/bitstamp"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/btse"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/coinbase"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/coinex"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/coinone"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/crypto"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/gateio"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/gemini"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/korbit"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/lbank"
	"bisonai.com/miko/node/pkg/websocketfetcher/providers/upbit"
	"github.com/stretchr/testify/assert"
)

var testFeeds = []common.Feed{
	{
		ID:         1,
		Name:       "binance-wss-BTC-USDT",
		Definition: json.RawMessage(`{"type": "wss", "provider": "binance", "base": "btc", "quote": "usdt"}`),
		ConfigID:   1,
	},
	{
		ID:         2,
		Name:       "coinbase-wss-ADA-USDT",
		Definition: json.RawMessage(`{"type": "wss", "provider": "coinbase", "base": "ada", "quote": "usdt"}`),
		ConfigID:   2,
	},
	{
		ID:         3,
		Name:       "coinone-wss-BTC-KRW",
		Definition: json.RawMessage(`{"type": "wss", "provider": "coinone", "base": "btc", "quote": "krw"}`),
		ConfigID:   3,
	},
	{
		ID:         4,
		Name:       "korbit-wss-BORA-KRW",
		Definition: json.RawMessage(`{"type": "wss", "provider": "korbit", "base": "bora", "quote": "krw"}`),
		ConfigID:   4,
	},
}

func TestGetWssFeedMap(t *testing.T) {
	feedMaps := common.GetWssFeedMap(testFeeds)
	if len(feedMaps) != 4 {
		t.Errorf("expected 4 feed maps, got %d", len(feedMaps))
	}

	for _, feed := range testFeeds {
		raw := strings.Split(feed.Name, "-")
		if len(raw) != 4 {
			t.Errorf("expected 4 parts, got %d", len(raw))
		}

		provider := strings.ToLower(raw[0])
		base := strings.ToUpper(raw[2])
		quote := strings.ToUpper(raw[3])
		combinedName := base + quote
		separatedName := base + "-" + quote

		if _, exists := feedMaps[provider]; !exists {
			t.Errorf("provider %s not found", provider)
		}
		if _, exists := feedMaps[provider].Combined[combinedName]; !exists {
			t.Errorf("combined feed %s not found", combinedName)
		}
		if _, exists := feedMaps[provider].Separated[separatedName]; !exists {
			t.Errorf("separated feed %s not found", separatedName)
		}
	}
}

func TestPriceStringToFloat64(t *testing.T) {
	price := "10000.123400"
	expected := 1000012340000.0
	result, err := common.PriceStringToFloat64(price)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("expected %f, got %f", expected, result)
	}
}

func TestMessageToStruct(t *testing.T) {
	t.Run("TestMessageToStructBinance", func(t *testing.T) {
		jsonStr := `{
			"e": "24hrMiniTicker",
			"E": 1672515782136,
			"s": "BNBBTC",
			"c": "0.0025",
			"o": "0.0010",
			"h": "0.0025",
			"l": "0.0010",
			"v": "10000",
			"q": "18"
		  }`

		var result map[string]any
		err := json.Unmarshal([]byte(jsonStr), &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		data, err := common.MessageToStruct[binance.MiniTicker](result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		assert.Equal(t, "BNBBTC", data.Symbol)
		assert.Equal(t, "0.0025", data.Price)
		assert.Equal(t, int64(1672515782136), data.EventTime)
	})

	t.Run("TestMessageToStructCoinbase", func(t *testing.T) {
		jsonStr := `{
			"type": "ticker",
			"sequence": 123456789,
			"product_id": "BTC-USD",
			"price": "50000.00",
			"open_24h": "48000.00",
			"volume_24h": "10000",
			"low_24h": "47000.00",
			"high_24h": "51000.00",
			"volume_30d": "300000",
			"best_bid": "49999.00",
			"best_bid_size": "0.5",
			"best_ask": "50001.00",
			"best_ask_size": "0.5",
			"side": "buy",
			"time": "2022-01-01T00:00:00Z",
			"trade_id": 1234,
			"last_size": "0.01"
		}`

		var result map[string]any
		err := json.Unmarshal([]byte(jsonStr), &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		data, err := common.MessageToStruct[coinbase.Ticker](result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		assert.Equal(t, "BTC-USD", data.ProductID)
		assert.Equal(t, "50000.00", data.Price)
		assert.Equal(t, "2022-01-01T00:00:00Z", data.Time)
	})

	t.Run("TestMessageToStructCoinone", func(t *testing.T) {
		jsonStr := `
			{
				"r":"DATA",
				"c":"TICKER",
				"d":{
				  "qc":"KRW",
				  "tc":"XRP",
				  "t":1693560378928,
				  "qv":"55827441390.8456",
				  "tv":"79912892.7741579",
				  "fi":"698.7",
				  "lo":"683.9",
				  "hi":"699.5",
				  "la":"687.9",
				  "vp":"100",
				  "abp":"688.3",
				  "abq":"84992.9448",
				  "bbp":"687.8",
				  "bbq":"13861.6179",
				  "i":"1693560378928001",
				  "yfi":"716.9",
				  "ylo":"690.4",
				  "yhi":"717.5",
				  "yla":"698.7",
				  "yqv":"41616318229.6505",
				  "ytv":"58248252.35151376"
				}
			  }
		`

		var result map[string]any
		err := json.Unmarshal([]byte(jsonStr), &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		data, err := common.MessageToStruct[coinone.Raw](result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		assert.Equal(t, "KRW", data.Data.QuoteCurrency)
		assert.Equal(t, "XRP", data.Data.TargetCurrency)
		assert.Equal(t, int64(1693560378928), data.Data.Timestamp)
		assert.Equal(t, "687.9", data.Data.Last)
	})

	t.Run("TestMessageToStructKorbit", func(t *testing.T) {
		jsonStr := `{
			"accessToken": null,
			"event": "korbit:push-ticker",
			"timestamp" : 1389678052000,
			"data":
			  {
				"channel": "ticker",
				"currency_pair": "btc_krw",
				"timestamp": 1558590089274,
				"last": "9198500.1235789",
				"open": "9500000.3445783",
				"bid": "9192500.4578344",
				"ask": "9198000.32148556",
				"low": "9171500.23785685",
				"high": "9599000.34876458",
				"volume": "1539.18571988",
				"change": "-301500.234578934"
			}
		  }`

		var result map[string]any
		err := json.Unmarshal([]byte(jsonStr), &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		data, err := common.MessageToStruct[korbit.Raw](result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "btc_krw", data.Data.CurrencyPair)
		assert.Equal(t, int64(1558590089274), data.Data.Timestamp)
		assert.Equal(t, "9198500.1235789", data.Data.Last)
	})

	t.Run("TestMessageToStructCryptoDotCom", func(t *testing.T) {
		jsonStr := `{
			"id": -1,
			"method": "subscribe",
			"code": 0,
			"result": {
			  "channel": "ticker",
			  "instrument_name": "ADA_USDT",
			  "subscription": "ticker.ADA_USDT",
			  "data": [
				{
				  "h": "0.45575",
				  "l": "0.44387",
				  "a": "0.44878",
				  "i": "ADA_USDT",
				  "v": "2900036",
				  "vv": "1303481.10",
				  "oi": "0",
				  "c": "0.0016",
				  "b": "0.44870",
				  "k": "0.44880",
				  "t": 1717223914135
				}
			  ]
			}
		  }`
		var result map[string]any
		err := json.Unmarshal([]byte(jsonStr), &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		data, err := common.MessageToStruct[crypto.Response](result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "ticker", data.Result.Channel)
		assert.Equal(t, "ADA_USDT", data.Result.InstrumentName)
		assert.Equal(t, int64(1717223914135), data.Result.Data[0].Timestamp)
		assert.Equal(t, "0.44878", *data.Result.Data[0].LastTradePrice)
	})

	t.Run("TestMessageToStructBtse", func(t *testing.T) {
		jsonStr := `{
			"topic": "tradeHistoryApi:ADA-USDT",
			"data": [
			  {
				"symbol": "ADA-USDT",
				"side": "SELL",
				"size": 122.4,
				"price": 0.44804,
				"tradeId": 62497538,
				"timestamp": 1717227427438
			  }
			]
		  }`

		var result map[string]any
		err := json.Unmarshal([]byte(jsonStr), &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		data, err := common.MessageToStruct[btse.Response](result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "ADA-USDT", data.Data[0].Symbol)
		assert.Equal(t, float64(0.44804), data.Data[0].Price)
		assert.Equal(t, int64(1717227427438), data.Data[0].Timestamp)
	})

	t.Run("TestMessageToStructBithumb", func(t *testing.T) {
		txJsonStr := `{
			"type": "transaction",
			"content": {
				"list": [
					{
						"symbol": "BTC_KRW",
						"buySellGb": "1",
						"contPrice": "10579000",
						"contQty": "0.01",
						"contAmt": "105790.00",
						"contDtm": "2020-01-29 12:24:18.830039",
						"updn": "dn"
					},
					{
						"symbol": "ETH_KRW",
						"buySellGb": "2",
						"contPrice": "200000",
						"contQty": "0.05",
						"contAmt": "10000.00",
						"contDtm": "2020-01-29 12:24:18.830039",
						"updn": "up"
					}
				]
			}
		}`

		tickerJsonStr := `{
			"type": "ticker",
			"content": {
				"symbol": "BTC_KRW",
				"tickType": "1H",
				"date": "20240601",
				"time": "171451",
				"openPrice": "1227",
				"closePrice": "1224",
				"lowPrice": "1223",
				"highPrice": "1230",
				"value": "22271989.6261801699999998",
				"volume": "18172.56112368162601626",
				"sellVolume": "10920.87235377",
				"buyVolume": "7251.68876991162601626",
				"prevClosePrice": "1211",
				"chgRate": "-0.24",
				"chgAmt": "-3",
				"volumePower": "66.4"
			}
		}`

		var txResult map[string]any
		err := json.Unmarshal([]byte(txJsonStr), &txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		txData, err := common.MessageToStruct[bithumb.TransactionResponse](txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "BTC_KRW", txData.Content.List[0].Symbol)
		assert.Equal(t, "10579000", txData.Content.List[0].ContPrice)
		assert.Equal(t, "2020-01-29 12:24:18.830039", txData.Content.List[0].ContDtm)

		var tickerResult map[string]any
		err = json.Unmarshal([]byte(tickerJsonStr), &tickerResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		tickerData, err := common.MessageToStruct[bithumb.TickerResponse](tickerResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "BTC_KRW", tickerData.Content.Symbol)
		assert.Equal(t, "20240601", tickerData.Content.Date)
		assert.Equal(t, "171451", tickerData.Content.Time)

	})

	t.Run("TestMessageToStructUpbit", func(t *testing.T) {
		jsonStr := `{
			"ty": "ticker",
			"cd": "KRW-BORA",
			"op": 204.1,
			"hp": 209.4,
			"lp": 204.1,
			"tp": 205.7,
			"pcp": 204.1,
			"c": "RISE",
			"cp": 1.6,
			"scp": 1.6,
			"cr": 0.0078392945,
			"scr": 0.0078392945,
			"tv": 772.15285594,
			"atv": 1.519251400033162e+07,
			"atv24h": 2.363356548175399e+07,
			"atp": 3.1378159174206986e+09,
			"atp24h": 4.862222332728324e+09,
			"td": "20240601",
			"ttm": "095356",
			"ttms": 1.717235636497e+12,
			"ab": "ASK",
			"aav": 8.55746037890694e+06,
			"abv": 6.63505362142468e+06,
			"h52wp": 346.4,
			"h52wdt": "2024-03-14",
			"l52wp": 140,
			"l52wdt": "2023-09-13",
			"ts": null,
			"ms": "ACTIVE",
			"msfi": null,
			"its": false,
			"dd": null,
			"mw": "NONE",
			"tms": 1.717235640279e+12,
			"st": "REALTIME",
			"sid": null
		  }`

		var txResult map[string]any
		err := json.Unmarshal([]byte(jsonStr), &txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		txData, err := common.MessageToStruct[upbit.Response](txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "KRW-BORA", txData.Code)
		assert.Equal(t, float64(205.7), txData.TradePrice)
		assert.Equal(t, int64(1.717235636497e+12), txData.TradeTimestamp)
	})

	t.Run("TestMessageToStructGateio", func(t *testing.T) {
		jsonStr := `{
			"time": 1717237882,
			"time_ms": 1717237882464,
			"channel": "spot.tickers",
			"event": "update",
			"result": {
			  "currency_pair": "ADA_USDT",
			  "last": "0.4479",
			  "lowest_ask": "0.448",
			  "highest_bid": "0.4479",
			  "change_percentage": "-0.1115",
			  "base_volume": "3532549.45",
			  "quote_volume": "1588551.077087",
			  "high_24h": "0.4558",
			  "low_24h": "0.4437"
			}
		  }`

		var txResult map[string]any
		err := json.Unmarshal([]byte(jsonStr), &txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		txData, err := common.MessageToStruct[gateio.Response](txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "spot.tickers", txData.Channel)
		assert.Equal(t, "ADA_USDT", txData.Result.CurrencyPair)
		assert.Equal(t, "0.4479", txData.Result.Last)
	})

	t.Run("TestMessageToStructCoinex", func(t *testing.T) {
		jsonStr := `{
    "method": "state.update",
    "data": {
        "state_list": [
            {
                "market": "LATUSDT",
                "last": "0.008157",
                "open": "0.008286",
                "close": "0.008157",
                "high": "0.008390",
                "low": "0.008106",
                "volume": "807714.49139758",
                "volume_sell": "286170.69645599",
                "volume_buy": "266161.23236408",
                "value": "6689.21644207",
                "period": 86400
            },
            {
                "market": "ELONUSDT",
                "last": "0.000000152823",
                "open": "0.000000158650",
                "close": "0.000000152823",
                "high": "0.000000159474",
                "low": "0.000000147026",
                "volume": "88014042237.15",
                "volume_sell": "11455578769.13",
                "volume_buy": "17047669612.10",
                "value": "13345.65122447",
                "period": 86400
            }
        ]
    },
    "id": null
}`
		var txResult map[string]any
		err := json.Unmarshal([]byte(jsonStr), &txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		txData, err := common.MessageToStruct[coinex.Response](txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "state.update", txData.Method)
		assert.Equal(t, "0.008157", txData.Data.StateList[0].Last)
	})

	t.Run("TestMessageToStructBitstamp", func(t *testing.T) {
		jsonStr := `{
			"channel": "live_trades_btcusd",
			"data": {
			  "id": 342188568,
			  "amount": 0.00054,
			  "amount_str": "0.00054000",
			  "price": 67703,
			  "price_str": "67703",
			  "type": 0,
			  "timestamp": "1717255024",
			  "microtimestamp": "1717255024403000",
			  "buy_order_id": 1754808523759616,
			  "sell_order_id": 1754808285728768
			},
			"event": "trade"
		  }`

		var txResult map[string]any
		err := json.Unmarshal([]byte(jsonStr), &txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		txData, err := common.MessageToStruct[bitstamp.TradeEvent](txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "live_trades_btcusd", txData.Channel)
		assert.Equal(t, "67703", txData.Data.PriceStr)

	})

	t.Run("TestMessageToStructGemini", func(t *testing.T) {
		jsonStr := `{
			"eventId": 1.713628958452305e+15,
			"events": [
			  {
				"amount": "0.00570297",
				"makerSide": "bid",
				"price": "67643.09",
				"symbol": "BTCUSD",
				"tid": "2.840140843076036e+15",
				"type": "trade"
			  }
			],
			"socket_sequence": 5,
			"timestamp": 1.717258316e+09,
			"timestampms": 1.717258316107e+12,
			"type": "update"
		  }`

		var txResult map[string]any
		err := json.Unmarshal([]byte(jsonStr), &txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		txData, err := common.MessageToStruct[gemini.Response](txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "BTCUSD", txData.Events[0].Symbol)
		assert.Equal(t, "67643.09", txData.Events[0].Price)
	})

	t.Run("TestMessageToStructLbank", func(t *testing.T) {
		jsonStr := `{
			"SERVER": "V2",
			"TS": "2024-06-02T00:40:21.527",
			"pair": "btc_usdt",
			"tick": {
				"change": 1.03,
				"cny": 490024.55,
				"dir": "sell",
				"high": 67988.17,
				"latest": 67668.93,
				"low": 66928.53,
				"to_cny": 7.24,
				"to_usd": 1,
				"turnover": 96477021.02,
				"usd": 67668.93,
				"vol": 1427.1578
			},
			"type": "tick"
		}`

		var txResult map[string]any
		err := json.Unmarshal([]byte(jsonStr), &txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		txData, err := common.MessageToStruct[lbank.Response](txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "btc_usdt", txData.Pair)
		assert.Equal(t, float64(67668.93), txData.Tick.Latest)

	})

	t.Run("TestMessageToStructBitget", func(t *testing.T) {
		jsonStr := `{
			"action": "snapshot",
			"arg": {
			  "channel": "ticker",
			  "instId": "ETHUSDT",
			  "instType": "sp"
			},
			"data": [
			  {
				"askSz": "1.4728",
				"baseVolume": "43525.8604",
				"bestAsk": "3802.73",
				"bestBid": "3802.72",
				"bidSz": "0.223",
				"chgUTC": "0.01084",
				"high24h": "3814.00",
				"instId": "ETHUSDT",
				"labeId": 0,
				"lastPR": "3802.72",
				"low24h": "3742.72",
				"open24h": "3798.77",
				"openUtc": "3761.93",
				"quoteVolume": "164693970.1506",
				"ts": "1719555625603"
			  }
			],
			"ts": 1.717262817225e+12
		  }`

		var txResult map[string]any
		err := json.Unmarshal([]byte(jsonStr), &txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		txData, err := common.MessageToStruct[bitget.Response](txResult)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		assert.Equal(t, "ETHUSDT", txData.Data[0].InstId)
		assert.Equal(t, "3802.72", txData.Data[0].Price)

	})
}
