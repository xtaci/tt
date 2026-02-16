package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// ════════════════════════════════════════════════════════════════════
// Terminal I/O — raw mode requires explicit \r\n
// ════════════════════════════════════════════════════════════════════

func emit(s string) {
	s = strings.ReplaceAll(s, "\n", "\r\n")
	os.Stdout.WriteString(s)
}

func emitf(f string, a ...any) { emit(fmt.Sprintf(f, a...)) }
func cls()                     { os.Stdout.WriteString("\033[2J\033[H") }
func bell()                    { os.Stdout.WriteString("\a") }
func hideCur()                 { os.Stdout.WriteString("\033[?25l") }
func showCur()                 { os.Stdout.WriteString("\033[?25h") }

// ════════════════════════════════════════════════════════════════════
// ANSI color constants
// ════════════════════════════════════════════════════════════════════

const (
	RST  = "\033[0m"
	BOLD = "\033[1m"
	DIM  = "\033[2m"
	REV  = "\033[7m"

	FgRed = "\033[91m"
	FgGrn = "\033[92m"
	FgYlw = "\033[93m"
	FgBlu = "\033[94m"
	FgCyn = "\033[96m"
	FgWht = "\033[97m"
	FgGry = "\033[90m"

	BgRed = "\033[41m"
)

// ════════════════════════════════════════════════════════════════════
// Box-drawing helpers (double-line Unicode frame)
// ════════════════════════════════════════════════════════════════════

const boxW = 72

func hTop() string   { return "╔" + strings.Repeat("═", boxW-2) + "╗" }
func hMid() string   { return "╠" + strings.Repeat("═", boxW-2) + "╣" }
func hBot() string   { return "╚" + strings.Repeat("═", boxW-2) + "╝" }
func hBlank() string { return "║" + strings.Repeat(" ", boxW-2) + "║" }

// hRow left-aligns content inside ║ ... ║, padding with spaces.
func hRow(s string) string {
	inner := boxW - 4
	pad := inner - vLen(s)
	if pad < 0 {
		pad = 0
	}
	return "║ " + s + strings.Repeat(" ", pad) + " ║"
}

// hCenter centres content inside ║ ... ║.
func hCenter(s string) string {
	inner := boxW - 4
	sl := vLen(s)
	if sl >= inner {
		return hRow(s)
	}
	left := (inner - sl) / 2
	right := inner - sl - left
	return "║ " + strings.Repeat(" ", left) + s + strings.Repeat(" ", right) + " ║"
}

// vLen returns visible length (ignoring ANSI escapes).
func vLen(s string) int {
	n, esc := 0, false
	for _, r := range s {
		if r == '\033' {
			esc = true
			continue
		}
		if esc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				esc = false
			}
			continue
		}
		n++
	}
	return n
}

// ════════════════════════════════════════════════════════════════════
// Lessons — modeled after the classic DOS TT
// ════════════════════════════════════════════════════════════════════

type Lesson struct {
	Name  string
	Lines []string
}

var lessons = []Lesson{
	{
		Name: "Lesson 1 — 主键区 Home Row",
		Lines: []string{
			"asdf jkl; asdf jkl; asdf jkl;",
			"fdsa ;lkj fdsa ;lkj fdsa ;lkj",
			"aa ss dd ff jj kk ll ;; aa ss",
			"fj fj dk dk sl sl a; a; fj dk",
			"add dad fad lad fall lass flask",
			"a sad lass; a glad dad; all fall",
			"ask a lad; add a flask; fall all",
			"glad dads fall; sad lads add all",
		},
	},
	{
		Name: "Lesson 2 — 上排键 Top Row",
		Lines: []string{
			"qwer tyui op qwer tyui op qwer",
			"rewq iuyt po rewq iuyt po rewq",
			"the quit wire poor your power",
			"quote quite tower rope ripe wipe",
			"we were quite proper true type",
			"your quiet power requires work",
			"write proper quotes quickly too",
			"ripe fruit requires quiet toil",
		},
	},
	{
		Name: "Lesson 3 — 下排键 Bottom Row",
		Lines: []string{
			"zxcv bnm, zxcv bnm, zxcv bnm,",
			"vcxz ,mnb vcxz ,mnb vcxz ,mnb",
			"can van ban man box mix cab zinc",
			"vim cave combine convex maximum",
			"move back next cabin became calm",
			"combine zinc vim cave move brave",
			"become maximum cabin convex back",
			"brave zinc move calm become next",
		},
	},
	{
		Name: "Lesson 4 — 全键盘混合",
		Lines: []string{
			"The quick brown fox jumps over the lazy dog.",
			"Pack my box with five dozen liquor jugs.",
			"How vexingly quick daft zebras jump!",
			"The five boxing wizards jump quickly.",
			"Jackdaws love my big sphinx of quartz.",
			"Crazy Frederick bought many very exquisite opal jewels.",
			"We promptly judged antique ivory buckles for the prize.",
			"A quick movement of the enemy will jeopardize gunboats.",
		},
	},
	{
		Name: "Lesson 5 — 数字与符号",
		Lines: []string{
			"1234567890 1234567890 1234567890",
			"a1b2c3d4 e5f6g7h8 i9j0 k1l2m3",
			"100 + 200 = 300; 55 * 3 = 165;",
			"price: $19.99; tax: 8.5%; total: $21.69",
			"email@host.com; http://example.org",
			"file_name.txt; path/to/dir; C:\\DOS",
			"(a + b) * c = d; [x] = {y};",
			"#include <stdio.h> /* comment */",
		},
	},
	{
		Name: "Lesson 6 — 编程代码",
		Lines: []string{
			"func main() {",
			`    fmt.Println("Hello, World!")`,
			"    for i := 0; i < 10; i++ {",
			"        sum += values[i]",
			"    }",
			"    if err != nil {",
			`        return fmt.Errorf("fail: %w", err)`,
			"}",
		},
	},
}

// ════════════════════════════════════════════════════════════════════
// Statistics
// ════════════════════════════════════════════════════════════════════

type Stats struct {
	Total   int
	Correct int
	Errors  int
	Elapsed time.Duration
}

func (st Stats) CPM() float64 {
	m := st.Elapsed.Minutes()
	if m <= 0 {
		return 0
	}
	return float64(st.Total) / m
}

func (st Stats) WPM() float64 { return st.CPM() / 5.0 }

func (st Stats) Accuracy() float64 {
	if st.Total == 0 {
		return 100
	}
	return float64(st.Correct) * 100 / float64(st.Total)
}

// ════════════════════════════════════════════════════════════════════
// Typing session  (line-by-line, like classic TT)
// ════════════════════════════════════════════════════════════════════

type Session struct {
	lesson    *Lesson
	lineIdx   int
	target    []rune // current line target
	typed     []rune // user input for current line
	errors    int    // error count for current line
	correct   int    // correct count for current line
	started   bool   // first key pressed?
	startTime time.Time
	lineStats []Stats // accumulated per-line
}

func newSession(l *Lesson) *Session {
	s := &Session{lesson: l}
	s.loadLine(0)
	return s
}

func (s *Session) loadLine(idx int) {
	s.lineIdx = idx
	s.target = []rune(s.lesson.Lines[idx])
	s.typed = nil
	s.errors = 0
	s.correct = 0
	s.started = false
}

func (s *Session) lineFinished() bool {
	return len(s.typed) >= len(s.target)
}

func (s *Session) allDone() bool {
	return s.lineIdx >= len(s.lesson.Lines)-1 && s.lineFinished()
}

func (s *Session) elapsed() time.Duration {
	if !s.started {
		return 0
	}
	return time.Since(s.startTime)
}

func (s *Session) addRune(r rune) {
	if s.lineFinished() {
		return
	}
	if !s.started {
		s.started = true
		s.startTime = time.Now()
	}
	pos := len(s.typed)
	if pos < len(s.target) {
		if r == s.target[pos] {
			s.correct++
		} else {
			s.errors++
			bell()
		}
	}
	s.typed = append(s.typed, r)
}

func (s *Session) backspace() {
	if len(s.typed) == 0 {
		return
	}
	pos := len(s.typed) - 1
	if pos < len(s.target) {
		if s.typed[pos] == s.target[pos] {
			s.correct--
		} else {
			s.errors--
		}
	}
	s.typed = s.typed[:pos]
}

func (s *Session) finishLine() Stats {
	el := s.elapsed()
	st := Stats{
		Total:   s.correct + s.errors,
		Correct: s.correct,
		Errors:  s.errors,
		Elapsed: el,
	}
	s.lineStats = append(s.lineStats, st)
	return st
}

func (s *Session) advanceLine() bool {
	if s.lineIdx+1 >= len(s.lesson.Lines) {
		return false
	}
	s.loadLine(s.lineIdx + 1)
	return true
}

func (s *Session) totalStats() Stats {
	var st Stats
	for _, ls := range s.lineStats {
		st.Total += ls.Total
		st.Correct += ls.Correct
		st.Errors += ls.Errors
		st.Elapsed += ls.Elapsed
	}
	return st
}

// ════════════════════════════════════════════════════════════════════
// Rendering
// ════════════════════════════════════════════════════════════════════

func renderMenu(sel int) {
	cls()
	var b strings.Builder
	b.WriteString(hTop() + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hCenter(BOLD+FgCyn+"╔╦╗╔╦╗  打 字 练 习"+RST) + "\n")
	b.WriteString(hCenter(BOLD+FgCyn+" ║  ║   Typing Tutor"+RST) + "\n")
	b.WriteString(hCenter(BOLD+FgCyn+" ╩  ╩   Go 复刻版"+RST) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hCenter(FgGry+"经典 DOS TT 风格 · 终端打字练习程序"+RST) + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(FgYlw+"选择课程:"+RST) + "\n")
	b.WriteString(hBlank() + "\n")
	for i, l := range lessons {
		marker := "  "
		color := FgWht
		if i == sel {
			marker = FgCyn + "▸ " + RST
			color = FgCyn + BOLD
		}
		b.WriteString(hRow(fmt.Sprintf("%s%s%d. %s%s", marker, color, i+1, l.Name, RST)) + "\n")
	}
	b.WriteString(hBlank() + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(FgGry+"↑↓ 选择 │ Enter 开始 │ Q 退出"+RST) + "\n")
	b.WriteString(hBot() + "\n")
	emit(b.String())
}

func renderTyping(s *Session) {
	cls()
	el := s.elapsed()
	mins := int(el.Minutes())
	secs := int(el.Seconds()) % 60
	total := s.correct + s.errors
	var cpm, acc float64
	if m := el.Minutes(); m > 0 {
		cpm = float64(total) / m
	}
	if total > 0 {
		acc = float64(s.correct) * 100 / float64(total)
	} else {
		acc = 100
	}

	lineInfo := fmt.Sprintf("第 %d/%d 行", s.lineIdx+1, len(s.lesson.Lines))
	statLine := fmt.Sprintf(
		"用时:%s%02d:%02d%s  速度:%s%.0f%s字/分  错误:%s%d%s  正确率:%s%.1f%%%s",
		FgYlw, mins, secs, RST,
		FgGrn, cpm, RST,
		FgRed, s.errors, RST,
		FgCyn, acc, RST,
	)

	// ── target line ──
	var targetBuf strings.Builder
	for _, r := range s.target {
		targetBuf.WriteRune(r)
	}

	// ── typed line with color coding ──
	var typedBuf strings.Builder
	for i, r := range s.typed {
		if i < len(s.target) {
			if r == s.target[i] {
				typedBuf.WriteString(FgGrn)
				typedBuf.WriteRune(s.target[i])
				typedBuf.WriteString(RST)
			} else {
				// show expected char on red background
				typedBuf.WriteString(BgRed + FgWht + BOLD)
				typedBuf.WriteRune(s.target[i])
				typedBuf.WriteString(RST)
			}
		}
	}
	// cursor: reverse-video on next expected char
	if !s.lineFinished() && len(s.typed) < len(s.target) {
		typedBuf.WriteString(REV)
		typedBuf.WriteRune(s.target[len(s.typed)])
		typedBuf.WriteString(RST)
	}

	var b strings.Builder
	b.WriteString(hTop() + "\n")
	b.WriteString(hCenter(BOLD+FgCyn+"TT — 打字练习"+RST) + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(fmt.Sprintf("%s%s%s    %s", BOLD+FgWht, s.lesson.Name, RST, lineInfo)) + "\n")
	b.WriteString(hRow(statLine) + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hRow(FgWht+BOLD+"目标: "+RST+FgWht+targetBuf.String()+RST) + "\n")
	b.WriteString(hRow(FgWht+BOLD+"输入: "+RST+typedBuf.String()) + "\n")
	b.WriteString(hBlank() + "\n")
	// progress bar
	progress := 0
	if len(s.target) > 0 {
		progress = len(s.typed) * 40 / len(s.target)
	}
	if progress > 40 {
		progress = 40
	}
	bar := FgGrn + strings.Repeat("█", progress) + FgGry + strings.Repeat("░", 40-progress) + RST
	pct := float64(len(s.typed)) * 100 / float64(len(s.target))
	b.WriteString(hRow(fmt.Sprintf("进度: [%s] %s%.0f%%%s", bar, FgYlw, pct, RST)) + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(FgGry+"Backspace=回退 │ ESC=菜单 │ Ctrl-C=退出"+RST) + "\n")
	b.WriteString(hBot() + "\n")
	emit(b.String())
}

func renderLineComplete(s *Session, st Stats) {
	cls()
	var b strings.Builder
	b.WriteString(hTop() + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hCenter(BOLD+FgGrn+"✓ 本行完成!"+RST) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(fmt.Sprintf("字符: %s%d%s   正确: %s%d%s   错误: %s%d%s",
		FgWht+BOLD, st.Total, RST,
		FgGrn+BOLD, st.Correct, RST,
		FgRed+BOLD, st.Errors, RST)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("用时: %s%.1f秒%s   速度: %s%.0f 字/分%s   正确率: %s%.1f%%%s",
		FgYlw, st.Elapsed.Seconds(), RST,
		FgGrn, st.CPM(), RST,
		FgCyn, st.Accuracy(), RST)) + "\n")
	b.WriteString(hMid() + "\n")
	if s.allDone() {
		b.WriteString(hRow(FgYlw+BOLD+"课程完成！按任意键查看总成绩..."+RST) + "\n")
	} else {
		b.WriteString(hRow(FgGry+"按任意键继续下一行..."+RST) + "\n")
	}
	b.WriteString(hBot() + "\n")
	emit(b.String())
}

func renderResults(s *Session) {
	cls()
	ts := s.totalStats()

	// grade
	grade := "F"
	switch {
	case ts.Accuracy() >= 98 && ts.CPM() >= 300:
		grade = "A+"
	case ts.Accuracy() >= 95 && ts.CPM() >= 250:
		grade = "A"
	case ts.Accuracy() >= 92 && ts.CPM() >= 200:
		grade = "B+"
	case ts.Accuracy() >= 90 && ts.CPM() >= 150:
		grade = "B"
	case ts.Accuracy() >= 85 && ts.CPM() >= 100:
		grade = "C"
	case ts.Accuracy() >= 80:
		grade = "D"
	}

	var b strings.Builder
	b.WriteString(hTop() + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hCenter(BOLD+FgCyn+"TT — 练习成绩单"+RST) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(fmt.Sprintf("课 程:  %s%s%s", FgWht+BOLD, s.lesson.Name, RST)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("总行数: %s%d%s", FgWht+BOLD, len(s.lesson.Lines), RST)) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hRow(fmt.Sprintf("总字符:  %s%d%s", FgWht+BOLD, ts.Total, RST)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("正 确:  %s%d%s", FgGrn+BOLD, ts.Correct, RST)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("错 误:  %s%d%s", FgRed+BOLD, ts.Errors, RST)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("总用时: %s%.1f 秒%s", FgYlw, ts.Elapsed.Seconds(), RST)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("速 度:  %s%.0f 字/分 (%.0f WPM)%s", FgGrn+BOLD, ts.CPM(), ts.WPM(), RST)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("正确率: %s%.1f%%%s", FgCyn+BOLD, ts.Accuracy(), RST)) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hCenter(BOLD+FgYlw+"评 级: "+grade+RST) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(FgGry+"R=重练 │ M=菜单 │ Q=退出"+RST) + "\n")
	b.WriteString(hBot() + "\n")
	emit(b.String())
}

// ════════════════════════════════════════════════════════════════════
// Keyboard input
// ════════════════════════════════════════════════════════════════════

const (
	evNone = iota
	evChar
	evBackspace
	evEnter
	evEscape
	evUp
	evDown
	evCtrlC
)

type keyEvent struct {
	kind int
	ch   rune
}

func readKey(buf []byte) keyEvent {
	n, err := os.Stdin.Read(buf)
	if err != nil || n == 0 {
		return keyEvent{kind: evNone}
	}
	switch {
	case n == 1 && buf[0] == 3:
		return keyEvent{kind: evCtrlC}
	case n == 1 && buf[0] == 27:
		return keyEvent{kind: evEscape}
	case n == 1 && (buf[0] == 127 || buf[0] == 8):
		return keyEvent{kind: evBackspace}
	case n == 1 && (buf[0] == '\r' || buf[0] == '\n'):
		return keyEvent{kind: evEnter}
	case n == 3 && buf[0] == 27 && buf[1] == '[':
		switch buf[2] {
		case 'A':
			return keyEvent{kind: evUp}
		case 'B':
			return keyEvent{kind: evDown}
		}
		return keyEvent{kind: evNone}
	case n >= 1 && buf[0] >= 32:
		// decode single UTF-8 rune
		r := rune(buf[0])
		if n >= 2 && buf[0]&0xE0 == 0xC0 {
			r = rune(buf[0]&0x1F)<<6 | rune(buf[1]&0x3F)
		} else if n >= 3 && buf[0]&0xF0 == 0xE0 {
			r = rune(buf[0]&0x0F)<<12 | rune(buf[1]&0x3F)<<6 | rune(buf[2]&0x3F)
		} else if n >= 4 && buf[0]&0xF8 == 0xF0 {
			r = rune(buf[0]&0x07)<<18 | rune(buf[1]&0x3F)<<12 | rune(buf[2]&0x3F)<<6 | rune(buf[3]&0x3F)
		}
		return keyEvent{kind: evChar, ch: r}
	}
	return keyEvent{kind: evNone}
}

// ════════════════════════════════════════════════════════════════════
// Main loop — state machine
// ════════════════════════════════════════════════════════════════════

const (
	stMenu = iota
	stTyping
	stLineEnd
	stResults
)

func run() error {
	fd := int(os.Stdin.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("无法进入终端原始模式: %w", err)
	}
	defer func() {
		showCur()
		term.Restore(fd, old)
	}()

	hideCur()
	buf := make([]byte, 16)
	sel := 0
	state := stMenu
	var sess *Session

	renderMenu(sel)

	for {
		k := readKey(buf)

		// Ctrl-C always quits
		if k.kind == evCtrlC {
			cls()
			emit("再见！\n")
			return nil
		}

		switch state {
		// ── Menu ──────────────────────────────────────────
		case stMenu:
			switch k.kind {
			case evUp:
				if sel > 0 {
					sel--
				}
				renderMenu(sel)
			case evDown:
				if sel < len(lessons)-1 {
					sel++
				}
				renderMenu(sel)
			case evEnter:
				sess = newSession(&lessons[sel])
				state = stTyping
				renderTyping(sess)
			case evChar:
				switch k.ch {
				case 'q', 'Q':
					cls()
					emit("再见！\n")
					return nil
				default:
					if k.ch >= '1' && k.ch <= rune('0'+len(lessons)) {
						sel = int(k.ch - '1')
						renderMenu(sel)
					}
				}
			case evEscape:
				cls()
				emit("再见！\n")
				return nil
			}

		// ── Typing ────────────────────────────────────────
		case stTyping:
			switch k.kind {
			case evEscape:
				state = stMenu
				renderMenu(sel)
			case evBackspace:
				sess.backspace()
				renderTyping(sess)
			case evChar:
				sess.addRune(k.ch)
				if sess.lineFinished() {
					st := sess.finishLine()
					state = stLineEnd
					renderLineComplete(sess, st)
				} else {
					renderTyping(sess)
				}
			}

		// ── Line complete ─────────────────────────────────
		case stLineEnd:
			// any key advances
			if k.kind == evNone {
				continue
			}
			if sess.allDone() {
				state = stResults
				renderResults(sess)
			} else {
				sess.advanceLine()
				state = stTyping
				renderTyping(sess)
			}

		// ── Results ───────────────────────────────────────
		case stResults:
			switch k.kind {
			case evChar:
				switch k.ch {
				case 'r', 'R':
					sess = newSession(sess.lesson)
					state = stTyping
					renderTyping(sess)
				case 'm', 'M':
					state = stMenu
					renderMenu(sel)
				case 'q', 'Q':
					cls()
					emit("再见！\n")
					return nil
				}
			case evEscape:
				state = stMenu
				renderMenu(sel)
			}
		}
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\r\n", err)
		os.Exit(1)
	}
}
