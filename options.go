package main

import (
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
)

// 변경된 부분: 체이닝 및 함수형 옵션 제거. 직접 struct literal 초기화 방식으로 사용.
// Ex) options.Opts{WaitResponse: true, DebugName: "ABC"}

type Opts struct {
	WaitResponse   bool
	Base           *cdp.Node // Base node for relative queries (especially useful for iframe)
	WaitVisible    bool
	Timeout        time.Duration
	LogTag         string
	DisableLog     bool
	ByTextUnstable bool
}

func (o *Opts) GetQueryOpts(sel *any, queryAll bool) []chromedp.QueryOption {
	// chromedp.QueryOption 생성
	var queryOpts []chromedp.QueryOption
	if o.Base != nil {
		queryOpts = append(queryOpts, chromedp.FromNode(o.Base))
	}
	if o.WaitVisible {
		queryOpts = append(queryOpts, chromedp.NodeVisible)
	}

	// sel, queryAll에 따라 ByQuery, ByQueryAll, ByNodeID 중 하나 선택
	if sel != nil {
		switch v := (*sel).(type) {
		case cdp.NodeID:
			queryOpts = append(queryOpts, chromedp.ByNodeID)
			*sel = []cdp.NodeID{v}
		case []cdp.NodeID:
			queryOpts = append(queryOpts, chromedp.ByNodeID)
		case string:
			queryOpts = append(queryOpts, lo.Ternary(queryAll, chromedp.ByQueryAll, chromedp.ByQuery))
		}
	}
	return queryOpts
}
