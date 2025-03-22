package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

const domain = "https://lms.ssu.ac.kr"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	id := os.Getenv("ID")
	pw := os.Getenv("PW")

	allocCtx, cancel := NewAllocator()
	defer cancel()

	browserCtx, _ := NewBrowser(allocCtx)

	tab, _ := NewTab(browserCtx)

	tab.GotoURL(domain + "/login")
	tab.Fill("#userid", id+"\t"+pw+"\n", &Opts{WaitResponse: true})
	tab.GotoURL(domain + "/mypage")
	iframe := tab.GetNode("iframe")

	time.Sleep(600 * time.Millisecond) // 최적화 완료
	var todoSubjects []*cdp.Node
	// TODO: 이거 함수로 만들기
	_, dur, _ := lo.WaitFor(func(_ int) bool {
		todoSubjects = lop.Map(
			tab.GetNodes(".xn-student-course-container", &Opts{Base: iframe, Timeout: 100 * time.Millisecond}),
			func(node *cdp.Node, i int) *cdp.Node {
				if tab.Text(" a", &Opts{Base: node, DisableLog: true}) != "0" {
					return node
				}
				return nil
			},
		)

		todoSubjects = lo.Compact(todoSubjects)

		return len(todoSubjects) > 0
	}, 10*time.Second, 1)
	fmt.Println("싸강있는 강의 파악. 대기시간:", dur)

	lo.ForEach(todoSubjects, func(subject *cdp.Node, _ int) {

		tab.Click(" button", &Opts{Base: subject})
		titles := tab.Texts(".xnsti-left:has(.video)", &Opts{Base: subject})

		if len(titles) == 0 {
			return
		}

		subjectTab, cancel := tab.OpenInNewTab(".xntc-count", &Opts{Base: subject, LogTag: fmt.Sprint(subject.NodeID, " 과목탭")})
		defer cancel()

		time.Sleep(5 * time.Second) // TODO: time.Sleep 없애기 여기 왜 에러
		iframe := subjectTab.GetNode("#tool_content")
		courseEntryNodeIds := lo.Filter(
			subjectTab.GetNodeIDs(".xnmb-module_item-wrapper:has(.readystream, .mp4) a", &Opts{Base: iframe}),
			func(nodeId cdp.NodeID, _ int) bool {
				return lo.Contains(titles, subjectTab.Text(nodeId))
			})

		lo.ForEach(courseEntryNodeIds, func(courseEntryNodeId cdp.NodeID, _ int) {

			courseTab, cancel := subjectTab.OpenInNewTab(courseEntryNodeId, &Opts{LogTag: fmt.Sprint(courseEntryNodeId, " 강의탭")})
			defer cancel()

			time.Sleep(4 * time.Second) // TODO: time.Sleep 없애기

			outerIframe := courseTab.GetNode("#tool_content")
			innerIframe := courseTab.GetNode("iframe", &Opts{Base: outerIframe})

			courseTab.Click("[title='재생']", &Opts{Base: innerIframe})

			lo.Times(2, func(_ int) any {
				time.Sleep(6 * time.Second) // TODO: time.Sleep 없애기
				if okBtnId := courseTab.GetNodeID(".confirm-ok-btn", &Opts{Base: innerIframe, WaitVisible: true, Timeout: 2 * time.Second}); okBtnId != 0 {
					courseTab.Click(okBtnId)
				}
				return nil
			})

			totalDuration := courseTab.Text(".xnvchp-info-duration>:nth-child(2)", &Opts{Base: outerIframe})
			currentDuration := strings.Split(courseTab.Text(".xnvc-progress-info-container>:nth-child(2)", &Opts{Base: outerIframe}), "(")[0]
			remainingDuration := parseKoreanDuration(totalDuration) - parseKoreanDuration(currentDuration)
			fmt.Printf("강의 남은시간: %v\n", remainingDuration)

			if IsProduction() {
				// TODO: Sleep도 로그 잘 띄워주는 함수로 만들자
				time.Sleep(remainingDuration)
			} else {
				time.Sleep(3 * time.Second)
			}
		})

	})
	fmt.Println("프로그램 종료")
}

var DefaultOpts = append(chromedp.DefaultExecAllocatorOptions[:],
	chromedp.Flag("mute-audio", true),                                 // 오디오 끄기
	chromedp.Flag("start-maximized", true),
	chromedp.Flag("headless", IsProduction()),                                // 직접 브라우저 보기
	chromedp.Flag("disable-blink-features", "AutomationControlled"), // 자동화 흔적 없애기 -> 구글 검색 뚫기
	chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.102 Safari/537.36"), // 자동화 흔적 없애기
	chromedp.NoSandbox,
	chromedp.Flag("disable-web-security", true),
	chromedp.Flag("disable-site-isolation-trials", true),
	chromedp.Flag("disable-beforeunload", true),
	chromedp.Flag("disable-popup-blocking", true),
	chromedp.Flag("disable-notifications", true),
	chromedp.Flag("disable-features", "IsolateOrigins,site-per-process"),
)

func NewAllocator() (context.Context, context.CancelFunc) {
	return chromedp.NewExecAllocator(context.Background(), DefaultOpts...)
}

func NewBrowser(allocatorCtx context.Context) (context.Context, context.CancelFunc) {
	return chromedp.NewContext(allocatorCtx)

}

func parseKoreanDuration(s string) time.Duration {
	re := regexp.MustCompile(`^\s*(?:(\d+)\s*시간)?\s*(?:(\d+)\s*분)?\s*(?:(\d+)\s*초)?\s*$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		sleepIfErr(fmt.Errorf("invalid duration string: %v", s))
	}
	var d time.Duration
	if matches[1] != "" {
		hours, err := strconv.Atoi(matches[1])
		sleepIfErr(err)
		d += time.Duration(hours) * time.Hour
	}
	if matches[2] != "" {
		minutes, err := strconv.Atoi(matches[2])
		sleepIfErr(err)
		d += time.Duration(minutes) * time.Minute
	}
	if matches[3] != "" {
		seconds, err := strconv.Atoi(matches[3])
		sleepIfErr(err)
		d += time.Duration(seconds) * time.Second
	}
	return d
}
