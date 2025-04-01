package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	lop "github.com/samber/lo/parallel"
)

const (
	green  = "\033[32m"
	red    = "\033[31m"
	reset  = "\033[0m"
	yellow = "\033[33m"
)

type Tab struct {
	BrowserCtx context.Context
	Ctx        context.Context
	Cancel     context.CancelFunc
}

func NewTab(browserCtx context.Context, opts ...*Opts) (*Tab, context.CancelFunc) {
	state := getState("NewTab", opts, nil, browserCtx)
	fmt.Println()

	ctx, cancel := chromedp.NewContext(browserCtx)
	newCancel := func() {
		cancel()
		if canLog(state) {
			fmt.Printf(`\n%v| %v [%v]`, time.Now(), "CloseTab", state.LogTag)
		}

	}

	preventAlert(ctx)

	// if canLog(state) {
	// 	fmt.Printf(`%v| %v [%v]\n`, time.Now(), "NewTab", state.LogTag)
	// }

	return &Tab{
		BrowserCtx: browserCtx,
		Ctx:        ctx,
		Cancel:     newCancel,
	}, newCancel
}

func (t *Tab) GotoURL(url string, opts ...*Opts) {
	state := getState("GotoURL", opts, nil, t.Ctx)

	_, state.Err = chromedp.RunResponse(state.Ctx, chromedp.Navigate(url))

	handleResult(state)
}

func (t *Tab) GetNode(sel any, opts ...*Opts) *cdp.Node {
	state := getState("GetNode", opts, &sel, t.Ctx)

	var nodes []*cdp.Node
	state.Err = chromedp.Run(state.Ctx, chromedp.Nodes(sel, &nodes, state.QueryOpts...))

	if handleResult(state) {
		if canLog(state) {
			fmt.Print(" ->", nodes[0].NodeID)
		}
		return nodes[0]
	}
	return nil
}

func (t *Tab) GetNodeID(sel any, opts ...*Opts) cdp.NodeID {
	state := getState("GetNodeID", opts, &sel, t.Ctx)

	var nodeids []cdp.NodeID
	state.Err = chromedp.Run(state.Ctx, chromedp.NodeIDs(sel, &nodeids, state.QueryOpts...))

	if handleResult(state) {
		if canLog(state) {
			fmt.Print(" ->", nodeids[0])
		}
		return nodeids[0]
	}
	return 0
}

func (t *Tab) Text(sel any, opts ...*Opts) string {
	state := getState("Text", opts, &sel, t.Ctx)

	var text string
	state.Err = chromedp.Run(state.Ctx, chromedp.Text(sel, &text, state.QueryOpts...))

	if handleResult(state) {
		if canLog(state) {
			fmt.Print(" ->", text)
		}
		return text
	}
	return ""
}

func (t *Tab) TextContent(sel any, opts ...*Opts) string {
	state := getState("TextContent", opts, &sel, t.Ctx)

	var text string
	state.Err = chromedp.Run(state.Ctx, chromedp.TextContent(sel, &text, state.QueryOpts...))

	if handleResult(state) {
		if canLog(state) {
			fmt.Println(" ->", text)
		}
		return text
	}
	return ""
}

func (t *Tab) GetAttribute(sel any, key string, opts ...*Opts) string {
	state := getState("GetAttribute", opts, &sel, t.Ctx)

	var value string
	state.Err = chromedp.Run(state.Ctx, chromedp.AttributeValue(sel, key, &value, nil, state.QueryOpts...))

	if handleResult(state) {
		if canLog(state) {
			fmt.Print(" ->", key, "=", value)
		}
		return value
	}
	return ""
}

func (t *Tab) SetAttribute(sel any, key, value string, opts ...*Opts) {
	state := getState("SetAttribute", opts, &sel, t.Ctx)

	state.Err = chromedp.Run(state.Ctx, chromedp.SetAttributeValue(sel, key, value, state.QueryOpts...))

	if handleResult(state) {
		if canLog(state) {
			fmt.Print(" ->", key, "=", value)
		}
	}
}

func (t *Tab) Fill(sel any, keys string, opts ...*Opts) {
	state := getState("Fill", opts, &sel, t.Ctx)

	action := chromedp.SendKeys(sel, keys, state.QueryOpts...)

	if state.WaitResponse {
		_, state.Err = chromedp.RunResponse(state.Ctx, action)
	} else {
		state.Err = chromedp.Run(state.Ctx, action)
	}

	if handleResult(state) {
		if canLog(state) {
			newStr := strings.ReplaceAll(keys, "\n", "\\n")
			newStr = strings.ReplaceAll(newStr, "\t", "\\t")
			fmt.Printf(` -> "%s"`, newStr)
		}
	}
}

func (t *Tab) Click(sel any, opts ...*Opts) {
	state := getState("Click", opts, &sel, t.Ctx)

	action := chromedp.Click(sel, state.QueryOpts...)

	if state.WaitResponse {
		_, state.Err = chromedp.RunResponse(state.Ctx, action)
	} else {
		state.Err = chromedp.Run(state.Ctx, action)
	}

	handleResult(state)
}

func (t *Tab) GetNodes(sel any, opts ...*Opts) []*cdp.Node {
	state := getState("GetNodes", opts, &sel, t.Ctx)

	var nodes []*cdp.Node
	state.Err = chromedp.Run(state.Ctx, chromedp.Nodes(sel, &nodes, state.QueryOpts...))

	if handleResult(state) {
		if canLog(state) {
			fmt.Print(" -> len=", len(nodes), ":", strings.Join(lop.Map(nodes, func(node *cdp.Node, i int) string {
				return fmt.Sprint(node.NodeID)
			}), ", "))
		}
		return nodes
	}
	return nil
}

func (t *Tab) GetNodeIDs(sel any, opts ...*Opts) []cdp.NodeID {
	state := getState("GetNodeIDs", opts, &sel, t.Ctx)

	var nodeids []cdp.NodeID
	state.Err = chromedp.Run(state.Ctx, chromedp.NodeIDs(sel, &nodeids, state.QueryOpts...))

	if handleResult(state) {
		if canLog(state) {
			fmt.Print(" -> len=", len(nodeids), ":", strings.Join(lop.Map(nodeids, func(nodeid cdp.NodeID, i int) string {
				return fmt.Sprint(nodeid)
			}), ", "))
		}
		return nodeids
	}
	return nil
}

func (t *Tab) Texts(sel any, opts ...*Opts) []string {
	state := getState("Texts", opts, &sel, t.Ctx)

	// GetNodeids
	var nodeids []cdp.NodeID
	state.Err = chromedp.Run(state.Ctx, chromedp.NodeIDs(sel, &nodeids, state.QueryOpts...))
	if state.Err != nil {
		handleResult(state)
		return nil
	}

	// Texts
	texts := lop.Map(nodeids, func(nodeid cdp.NodeID, i int) string {
		var text string
		state.Err = chromedp.Run(state.Ctx, chromedp.Text([]cdp.NodeID{nodeid}, &text, chromedp.ByNodeID))
		return text
	})

	if handleResult(state) {
		if canLog(state) {
			fmt.Print(" -> len=", len(texts), ":", strings.Join(lop.Map(texts, func(text string, _ int) string { return `"` + text + `"` }), ", "))
		}
		return texts
	}
	return nil
}

func (t *Tab) TextContents(sel any, opts ...*Opts) []string {
	state := getState("TextContents", opts, &sel, t.Ctx)

	// GetNodeids
	var nodeids []cdp.NodeID
	state.Err = chromedp.Run(state.Ctx, chromedp.NodeIDs(sel, &nodeids, state.QueryOpts...))
	if state.Err != nil {
		handleResult(state)
		return nil
	}

	// Texts
	textContents := lop.Map(nodeids, func(nodeid cdp.NodeID, i int) string {
		var textContent string
		state.Err = chromedp.Run(state.Ctx, chromedp.TextContent([]cdp.NodeID{nodeid}, &textContent, chromedp.ByNodeID))
		return textContent
	})

	if handleResult(state) {
		if canLog(state) {
			fmt.Print(" -> len=", len(textContents), ":", strings.Join(lop.Map(textContents, func(textContent string, _ int) string { return `"` + textContent + `"` }), ", "))
		}
		return textContents
	}
	return nil
}

func (t *Tab) Evaluate(js string, opts ...*Opts) (any, error) {
	state := getState("Eval", opts, nil, t.Ctx)

	var result any
	state.Err = chromedp.Run(state.Ctx, chromedp.Evaluate(js, &result))

	if handleResult(state) {
		if canLog(state) {
			fmt.Print(" ->", result)
		}
		return result, nil
	}
	return "", state.Err
}

func (t *Tab) OpenInNewTab(sel any, opts ...*Opts) (*Tab, context.CancelFunc) {
	state := getState("OpenInNewTab", opts, &sel, t.Ctx)

	ch := chromedp.WaitNewTarget(state.Ctx, func(info *target.Info) bool {
		return info.URL != ""
	})

	// t.SetAttribute(sel, "target", "_blank", opts...)
	state.Err = chromedp.Run(state.Ctx, chromedp.SetAttributeValue(sel, "target", "_blank", state.QueryOpts...))
	if state.Err != nil {
		handleResult(state)
		return nil, nil
	}

	// t.Fill(sel, "\n", opts...)
	state.Err = chromedp.Run(state.Ctx, chromedp.Click(sel, state.QueryOpts...))
	if state.Err != nil {
		handleResult(state)
		return nil, nil
	}

	targetID := <-ch
	// TODO: 여기서 state.Ctx 써야할지 t.Ctx 써야할지
	newTabCtx, cancel := chromedp.NewContext(t.Ctx, chromedp.WithTargetID(targetID))
	preventAlert(newTabCtx)

	newCancel := func() {
		cancel()
		if canLog(state) {
			fmt.Println()
			fmt.Printf(`%v| %s [%v] targetID: %v`, time.Now(), "CloseTab", state.LogTag, targetID)
		}

	}

	if handleResult(state) {
		if canLog(state) {
			fmt.Print(" -> targetID: ", targetID)
		}

		return &Tab{
			BrowserCtx: t.BrowserCtx,
			Ctx:        newTabCtx,
			Cancel:     newCancel,
		}, newCancel
	}

	return nil, nil
}

func (t *Tab) OpenFrameInNewTab(sel any, opts ...*Opts) (*Tab, context.CancelFunc) {
	state := getState("OpenFrameInNewTab", opts, &sel, t.Ctx)

	// ch := chromedp.WaitNewTarget(state.Ctx, func(info *target.Info) bool {
	// 	return info.URL != ""
	// })

	url := t.GetAttribute(sel, "src", opts...)

	// targetID := <-ch
	newTab, cancel := NewTab(t.Ctx)
	newTab.GotoURL(url)

	if handleResult(state) {
		return newTab, cancel
	}

	return nil, nil
}

func sleepIfErr(err error) {
	if err != nil {
		fmt.Println("\n", red, err, reset)
		time.Sleep(time.Hour)
	}
}

func preventAlert(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev any) {
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			go func() { chromedp.Run(ctx, page.HandleJavaScriptDialog(true)) }()
		}
	})
}

// returns isSuccessful
func handleResult(state *State) bool {
	if state.Cancel != nil {
		defer state.Cancel()
	}

	if state.Err != nil && state.Ctx.Err() != context.DeadlineExceeded {
		sleepIfErr(state.Err)
	}

	if canLog(state) {
		if state.Err != nil {
			fmt.Printf("%sTimeout (+%v)%s", yellow, time.Since(state.StartTime), reset)
		} else {
			fmt.Printf("%sSuccess (+%v)%s", green, time.Since(state.StartTime), reset)
		}
	}

	return state.Err == nil
}

func getState(funcName string, opts []*Opts, sel *any, tapCtx context.Context) *State {
	if len(opts) == 0 {
		opts = append(opts, &Opts{})
	}

	state := &State{Opts: *opts[0]}
	state.FuncName = funcName
	state.Ctx = tapCtx

	state.QueryOpts = opts[0].GetQueryOpts(sel, strings.HasSuffix(funcName, "s"))

	// Timeout 설정이 있으면 context.WithTimeout 생성
	if state.Timeout != 0 {
		state.Ctx, state.Cancel = context.WithTimeout(tapCtx, state.Timeout)
	}

	// 로깅 처리
	if canLog(state) {
		state.StartTime = time.Now()
		fmt.Println()
		fmt.Printf(`%v | %v <%s>`, state.StartTime.Format("15:04:05.000"), funcName, state.LogTag)
		if sel != nil {
			fmt.Printf(`"%v" ... `, *sel)
		}
	}

	return state
}

func canLog(state *State) bool {
	return !IsProduction() && !state.DisableLog
}

func IsProduction() bool {
	return os.Getenv("D") != "1"
}

type State struct {
	Opts
	FuncName  string
	StartTime time.Time
	QueryOpts []chromedp.QueryOption
	Ctx       context.Context
	Cancel    context.CancelFunc
	Err       error
}
