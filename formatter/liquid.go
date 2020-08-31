package formatter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/exchangedataset/streamcommons"
	"github.com/exchangedataset/streamcommons/formatter/jsondef"
	"github.com/exchangedataset/streamcommons/jsonstructs"
)

// liquidFormatter formats messages from Liquid to json format.
type liquidFormatter struct {
}

// FormatStart returns empty slice.
func (f *liquidFormatter) FormatStart(urlStr string) (formatted []StartReturn, err error) {
	return nil, nil
}

// FormatMessage formats raw messages from Liquid server to json format.
func (f *liquidFormatter) FormatMessage(channel string, line []byte) (ret [][]byte, err error) {
	r := new(jsonstructs.LiquidMessageRoot)
	serr := json.Unmarshal(line, r)
	if serr != nil {
		err = fmt.Errorf("FormatMessage: root: %v", serr)
		return
	}
	if r.Event == jsonstructs.LiquidEventSubscriptionSucceeded {
		if strings.HasPrefix(channel, streamcommons.LiquidChannelPrefixLaddersCash) {
			return [][]byte{jsondef.TypeDefLiquidPriceLaddersCash}, nil
		} else if strings.HasPrefix(channel, streamcommons.LiquidChannelPrefixExecutionsCash) {
			return [][]byte{jsondef.TypeDefLiquidExecutionsCash}, nil
		} else {
			return nil, fmt.Errorf("FormatMessage: channel not supported: %v", channel)
		}
	}
	if strings.HasPrefix(channel, streamcommons.LiquidChannelPrefixLaddersCash) {
		// `true` is ask
		side := strings.HasSuffix(channel, "sell")
		orderbook := make([][]string, 0, 100)
		serr = json.Unmarshal(r.Data, &orderbook)
		if serr != nil {
			return nil, fmt.Errorf("FormatMessage: price ladder: %v", orderbook)
		}
		ret = make([][]byte, len(orderbook))
		for i, memOrder := range orderbook {
			price, serr := strconv.ParseFloat(memOrder[0], 64)
			if serr != nil {
				return nil, fmt.Errorf("FormatMessage: price: %v", serr)
			}
			quantity, serr := strconv.ParseFloat(memOrder[1], 64)
			if serr != nil {
				return nil, fmt.Errorf("FormatMessage: quantity: %v", serr)
			}
			order := new(jsondef.LiquidPriceLaddersCash)
			order.Price = price
			if side {
				order.Size = -quantity
			} else {
				order.Size = quantity
			}
			om, serr := json.Marshal(order)
			if serr != nil {
				return nil, fmt.Errorf("FormatMessage: order marshal: %v", serr)
			}
			ret[i] = om
		}
		return ret, nil
	} else if strings.HasPrefix(channel, streamcommons.LiquidChannelPrefixExecutionsCash) {
		execution := new(jsonstructs.LiquidExecution)
		serr = json.Unmarshal(r.Data, execution)
		if serr != nil {
			return nil, fmt.Errorf("FormatMessage: execution: %v", serr)
		}
		formatted := new(jsondef.LiquidExecutionsCash)
		createdAt := time.Unix(int64(execution.CreatedAt), 0)
		formatted.CreatedAt = strconv.FormatInt(createdAt.UnixNano(), 10)
		formatted.ID = execution.ID
		formatted.Symbol = channel[len(streamcommons.LiquidChannelPrefixExecutionsCash):]
		formatted.Price = execution.Price
		if execution.TakerSide == "sell" {
			formatted.Size = -execution.Quantity
		} else {
			formatted.Size = execution.Quantity
		}
		fm, serr := json.Marshal(formatted)
		if serr != nil {
			return nil, fmt.Errorf("FormatMessage: formatted: %v", serr)
		}
		ret = [][]byte{fm}
	}
	return nil, fmt.Errorf("FormatMessage: line not supported")
}