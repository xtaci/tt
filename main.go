package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
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
func cls() {
	os.Stdout.WriteString("\033[2J\033[H")
	// Fill the entire terminal background with TT deep blue
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w, h = 80, 25
	}
	for i := 0; i < h; i++ {
		os.Stdout.WriteString(fmt.Sprintf("\033[%d;1H"+TTBg+strings.Repeat(" ", w), i+1))
	}
	os.Stdout.WriteString("\033[H")
}
func bell()    { os.Stdout.WriteString("\a") }
func hideCur() { os.Stdout.WriteString("\033[?25l") }
func showCur() { os.Stdout.WriteString("\033[?25h") }

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
	FgBlk = "\033[30m"

	BgRed  = "\033[41m"
	BgBlu  = "\033[44m"
	BgDkBl = "\033[48;5;17m" // Deep dark blue — classic DOS TT
	BgCyn  = "\033[46m"
	BgBlk  = "\033[40m"
	BgGry  = "\033[100m"

	// Classic DOS TT palette aliases
	TTBg     = "\033[48;5;17m" // Deep blue background
	TTFg     = "\033[97m"      // Bright white foreground
	TTTitle  = "\033[93m"      // Yellow titles
	TTBorder = "\033[96m"      // Cyan borders
	TTHilite = "\033[92m"      // Green highlights
	TTErr    = "\033[91m"      // Red errors
	TTDim    = "\033[37m"      // Dimmed text
)

// ════════════════════════════════════════════════════════════════════
// Box-drawing helpers (double-line Unicode frame)
// ════════════════════════════════════════════════════════════════════

const boxW = 72

func hTop() string { return TTBg + TTBorder + "╔" + strings.Repeat("═", boxW-2) + "╗" + RST }
func hMid() string { return TTBg + TTBorder + "╠" + strings.Repeat("═", boxW-2) + "╣" + RST }
func hBot() string { return TTBg + TTBorder + "╚" + strings.Repeat("═", boxW-2) + "╝" + RST }
func hBlank() string {
	return TTBg + TTBorder + "║" + TTFg + strings.Repeat(" ", boxW-2) + TTBorder + "║" + RST
}

// hRow left-aligns content inside ║ ... ║, padding with spaces.
func hRow(s string) string {
	inner := boxW - 4
	pad := inner - vLen(s)
	if pad < 0 {
		pad = 0
	}
	return TTBg + TTBorder + "║ " + TTFg + s + TTFg + strings.Repeat(" ", pad) + TTBorder + " ║" + RST
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
	return TTBg + TTBorder + "║ " + TTFg + strings.Repeat(" ", left) + s + strings.Repeat(" ", right) + TTBorder + " ║" + RST
}

// vLen returns the visible display width of s in terminal columns,
// correctly handling ANSI escape sequences and East Asian wide characters.
func vLen(s string) int {
	w, esc := 0, false
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
		w += runeWidth(r)
	}
	return w
}

// runeWidth returns the display column width of a rune in a terminal.
func runeWidth(r rune) int {
	if isWide(r) {
		return 2
	}
	return 1
}

// isWide reports whether r is an East Asian wide character
// that occupies 2 columns in a terminal.
func isWide(r rune) bool {
	switch {
	case r >= 0x1100 && r <= 0x115F: // Hangul Jamo
		return true
	case r >= 0x2E80 && r <= 0x303E: // CJK Radicals, Kangxi, Symbols
		return true
	case r >= 0x3041 && r <= 0x33BF: // Hiragana, Katakana, Bopomofo, Compat
		return true
	case r >= 0x3400 && r <= 0x4DBF: // CJK Extension A
		return true
	case r >= 0x4E00 && r <= 0xA4CF: // CJK Unified Ideographs, Yi
		return true
	case r >= 0xAC00 && r <= 0xD7AF: // Hangul Syllables
		return true
	case r >= 0xF900 && r <= 0xFAFF: // CJK Compatibility Ideographs
		return true
	case r >= 0xFE30 && r <= 0xFE6F: // CJK Compatibility Forms
		return true
	case r >= 0xFF01 && r <= 0xFF60: // Fullwidth Forms
		return true
	case r >= 0xFFE0 && r <= 0xFFE6: // Fullwidth Signs
		return true
	case r >= 0x20000 && r <= 0x2FA1F: // CJK Extension B+ & Compat Suppl
		return true
	case r >= 0x1F300 && r <= 0x1FAFF: // Emoji Symbols & Pictographs
		return true
	case r >= 0x2600 && r <= 0x27BF: // Misc Symbols & Dingbats
		return true
	case r >= 0x231A && r <= 0x23FF: // Misc Technical (some emoji)
		return true
	}
	return false
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
		Name: "Lesson 1 — Home Row",
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
		Name: "Lesson 2 — Top Row",
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
		Name: "Lesson 3 — Bottom Row",
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
		Name: "Lesson 4 — Full Keyboard",
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
		Name: "Lesson 5 — Numbers & Symbols",
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
		Name: "Lesson 6 — Programming",
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
	nItems := len(lessons) + 1 // +1 for Space Invaders
	var b strings.Builder
	b.WriteString(hTop() + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hCenter(BOLD+TTTitle+"╔╦╗╔╦╗  Typing Tutor"+RST+TTBg) + "\n")
	b.WriteString(hCenter(BOLD+TTTitle+" ║  ║   DOS TT Clone"+RST+TTBg) + "\n")
	b.WriteString(hCenter(BOLD+TTTitle+" ╩  ╩   in Golang   "+RST+TTBg) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hCenter(TTDim+"Classic DOS TT Style Terminal Typing Practice"+RST+TTBg) + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(TTTitle+"Select Lesson:"+RST+TTBg) + "\n")
	b.WriteString(hBlank() + "\n")
	for i, l := range lessons {
		marker := "  "
		color := TTFg
		if i == sel {
			marker = FgCyn + "▸ " + RST + TTBg
			color = FgCyn + BOLD
		}
		b.WriteString(hRow(fmt.Sprintf("%s%s%d. %s%s", marker, color, i+1, l.Name, RST+TTBg)) + "\n")
	}
	// Space Invaders entry
	siIdx := len(lessons)
	marker := "  "
	color := TTFg
	if sel == siIdx {
		marker = FgCyn + "▸ " + RST + TTBg
		color = FgCyn + BOLD
	}
	b.WriteString(hRow(fmt.Sprintf("%s%s%d. Space Invaders -- Typing Game%s", marker, color, siIdx+1, RST+TTBg)) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(TTDim+"Up/Down Select │ Enter Start │ Q Quit"+RST+TTBg) + "\n")
	b.WriteString(hBot() + "\n")
	_ = nItems
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

	lineInfo := fmt.Sprintf("Line %d/%d", s.lineIdx+1, len(s.lesson.Lines))
	statLine := fmt.Sprintf(
		"Time:%s%02d:%02d%s  Speed:%s%.0f%sCPM  Errors:%s%d%s  Accuracy:%s%.1f%%%s",
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
	b.WriteString(hCenter(BOLD+TTTitle+"TT — Typing Practice"+RST+TTBg) + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(fmt.Sprintf("%s%s%s    %s", BOLD+TTFg, s.lesson.Name, RST+TTBg, lineInfo)) + "\n")
	b.WriteString(hRow(statLine) + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hRow(FgWht+BOLD+"Target: "+RST+TTBg+FgWht+targetBuf.String()+RST+TTBg) + "\n")
	b.WriteString(hRow(FgWht+BOLD+"Input:  "+RST+TTBg+typedBuf.String()+RST+TTBg) + "\n")
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
	b.WriteString(hRow(fmt.Sprintf("Progress: [%s] %s%.0f%%%s", bar, FgYlw, pct, RST+TTBg)) + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(TTDim+"Backspace=Delete │ ESC=Menu │ Ctrl-C=Quit"+RST+TTBg) + "\n")
	b.WriteString(hBot() + "\n")
	emit(b.String())
}

func renderLineComplete(s *Session, st Stats) {
	cls()
	var b strings.Builder
	b.WriteString(hTop() + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hCenter(BOLD+TTHilite+"✓ Line Complete!"+RST+TTBg) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(fmt.Sprintf("Chars: %s%d%s   Correct: %s%d%s   Errors: %s%d%s",
		TTFg+BOLD, st.Total, RST+TTBg,
		FgGrn+BOLD, st.Correct, RST+TTBg,
		FgRed+BOLD, st.Errors, RST+TTBg)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("Time: %s%.1fs%s   Speed: %s%.0f CPM%s   Accuracy: %s%.1f%%%s",
		FgYlw, st.Elapsed.Seconds(), RST+TTBg,
		FgGrn, st.CPM(), RST+TTBg,
		FgCyn, st.Accuracy(), RST+TTBg)) + "\n")
	b.WriteString(hMid() + "\n")
	if s.allDone() {
		b.WriteString(hRow(TTTitle+BOLD+"Lesson complete! Press any key for results..."+RST+TTBg) + "\n")
	} else {
		b.WriteString(hRow(TTDim+"Press any key for next line..."+RST+TTBg) + "\n")
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
	b.WriteString(hCenter(BOLD+TTTitle+"TT — Score Report"+RST+TTBg) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(fmt.Sprintf("Lesson:   %s%s%s", TTFg+BOLD, s.lesson.Name, RST+TTBg)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("Lines:    %s%d%s", TTFg+BOLD, len(s.lesson.Lines), RST+TTBg)) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hRow(fmt.Sprintf("Chars:    %s%d%s", TTFg+BOLD, ts.Total, RST+TTBg)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("Correct:  %s%d%s", FgGrn+BOLD, ts.Correct, RST+TTBg)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("Errors:   %s%d%s", FgRed+BOLD, ts.Errors, RST+TTBg)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("Time:     %s%.1fs%s", FgYlw, ts.Elapsed.Seconds(), RST+TTBg)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("Speed:    %s%.0f CPM (%.0f WPM)%s", FgGrn+BOLD, ts.CPM(), ts.WPM(), RST+TTBg)) + "\n")
	b.WriteString(hRow(fmt.Sprintf("Accuracy: %s%.1f%%%s", FgCyn+BOLD, ts.Accuracy(), RST+TTBg)) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hCenter(BOLD+TTTitle+"Grade: "+grade+RST+TTBg) + "\n")
	b.WriteString(hBlank() + "\n")
	b.WriteString(hMid() + "\n")
	b.WriteString(hRow(TTDim+"R=Retry │ M=Menu │ Q=Quit"+RST+TTBg) + "\n")
	b.WriteString(hBot() + "\n")
	emit(b.String())
}

// ════════════════════════════════════════════════════════════════════
// Space Invaders — Typing Game (classic TT feature)
// ════════════════════════════════════════════════════════════════════

const (
	siFieldW = 60 // play field width
	siFieldH = 18 // play field height (rows aliens can occupy)
)

// Invader is a single falling alien with a letter on it.
type Invader struct {
	ch   rune    // letter to type
	x    int     // column position (0-based)
	y    float64 // row position (fractional, rendered as int)
	dead bool    // destroyed?
}

// SpaceGame holds the full game state.
type SpaceGame struct {
	invaders   []Invader
	score      int
	lives      int
	level      int
	missed     int
	hits       int
	speed      float64 // rows per tick
	spawnRate  int     // ticks between spawns
	tick       int
	lastRender time.Time
	tickRate   time.Duration
	gameOver   bool
}

func newSpaceGame() *SpaceGame {
	g := &SpaceGame{
		lives:     3,
		level:     1,
		speed:     0.3,
		spawnRate: 8,
		tickRate:  120 * time.Millisecond,
	}
	return g
}

func (g *SpaceGame) spawnInvader() {
	// pick a random lowercase letter
	letters := "abcdefghijklmnopqrstuvwxyz"
	ch := rune(letters[rand.Intn(len(letters))])
	x := rand.Intn(siFieldW-4) + 2
	g.invaders = append(g.invaders, Invader{ch: ch, x: x, y: 0})
}

func (g *SpaceGame) update() {
	if g.gameOver {
		return
	}
	g.tick++

	// spawn new invaders
	if g.tick%g.spawnRate == 0 {
		g.spawnInvader()
	}

	// move invaders down
	alive := g.invaders[:0]
	for i := range g.invaders {
		inv := &g.invaders[i]
		if inv.dead {
			continue
		}
		inv.y += g.speed
		if int(inv.y) >= siFieldH {
			// reached bottom — lost a life
			g.lives--
			g.missed++
			bell()
			if g.lives <= 0 {
				g.gameOver = true
			}
			continue
		}
		alive = append(alive, *inv)
	}
	g.invaders = alive

	// level up every 15 hits
	newLevel := g.hits/15 + 1
	if newLevel > g.level {
		g.level = newLevel
		g.speed += 0.08
		if g.spawnRate > 3 {
			g.spawnRate--
		}
	}
}

func (g *SpaceGame) tryShoot(ch rune) bool {
	// find the lowest (closest to bottom) invader with this letter
	bestIdx := -1
	bestY := -1.0
	for i, inv := range g.invaders {
		if !inv.dead && inv.ch == ch {
			if inv.y > bestY {
				bestY = inv.y
				bestIdx = i
			}
		}
	}
	if bestIdx >= 0 {
		g.invaders[bestIdx].dead = true
		g.score += 10 * g.level
		g.hits++
		return true
	}
	return false
}

func renderSpaceGame(g *SpaceGame) {
	cls()
	var b strings.Builder

	b.WriteString(hTop() + "\n")
	b.WriteString(hCenter(BOLD+TTTitle+"** SPACE INVADERS -- Type to Shoot! **"+RST+TTBg) + "\n")
	b.WriteString(hMid() + "\n")

	// Status bar
	livesStr := strings.Repeat("* ", g.lives) + strings.Repeat("  ", 3-g.lives)
	b.WriteString(hRow(fmt.Sprintf(
		"Score:%s%5d%s  Level:%s%d%s  Lives:%s%s%s  Hits:%s%d%s",
		FgYlw+BOLD, g.score, RST+TTBg,
		FgCyn+BOLD, g.level, RST+TTBg,
		FgRed+BOLD, livesStr, RST+TTBg,
		FgGrn+BOLD, g.hits, RST+TTBg,
	)) + "\n")
	b.WriteString(hMid() + "\n")

	// Build the play field
	field := make([][]rune, siFieldH)
	for r := 0; r < siFieldH; r++ {
		field[r] = make([]rune, siFieldW)
		for c := 0; c < siFieldW; c++ {
			field[r][c] = ' '
		}
	}

	// Place invaders on the field
	for _, inv := range g.invaders {
		if inv.dead {
			continue
		}
		row := int(inv.y)
		if row >= 0 && row < siFieldH && inv.x >= 0 && inv.x < siFieldW {
			field[row][inv.x] = inv.ch
		}
	}

	// Render field rows
	for r := 0; r < siFieldH; r++ {
		var row strings.Builder
		for c := 0; c < siFieldW; c++ {
			ch := field[r][c]
			if ch != ' ' {
				// Color aliens by proximity: green at top, yellow mid, red near bottom
				color := FgGrn
				if r > siFieldH*2/3 {
					color = FgRed + BOLD
				} else if r > siFieldH/3 {
					color = FgYlw + BOLD
				} else {
					color = FgGrn + BOLD
				}
				row.WriteString(color)
				row.WriteRune(ch)
				row.WriteString(RST + TTBg)
			} else {
				// Starfield: occasional dim dots
				if (r*siFieldW+c)%37 == 0 {
					row.WriteString(DIM + "·" + RST + TTBg)
				} else {
					row.WriteRune(' ')
				}
			}
		}
		// Pad to fit inside the box (boxW-4 inner width, field is siFieldW)
		padLeft := (boxW - 4 - siFieldW) / 2
		padRight := boxW - 4 - siFieldW - padLeft
		if padLeft < 0 {
			padLeft = 0
		}
		if padRight < 0 {
			padRight = 0
		}
		b.WriteString(TTBg + TTBorder + "║ " + TTFg +
			strings.Repeat(" ", padLeft) +
			row.String() +
			strings.Repeat(" ", padRight) +
			TTBorder + " ║" + RST + "\n")
	}

	// Cannon at bottom
	cannonPad := (boxW - 4 - siFieldW) / 2
	if cannonPad < 0 {
		cannonPad = 0
	}
	cannon := strings.Repeat(" ", siFieldW/2-1) + FgCyn + BOLD + "▲" + RST + TTBg + strings.Repeat(" ", siFieldW/2)
	b.WriteString(TTBg + TTBorder + "║ " + TTFg +
		strings.Repeat(" ", cannonPad) + cannon +
		strings.Repeat(" ", (boxW-4-siFieldW)-cannonPad) +
		TTBorder + " ║" + RST + "\n")

	b.WriteString(hMid() + "\n")
	if g.gameOver {
		b.WriteString(hCenter(BOLD+FgRed+"GAME OVER!"+RST+TTBg) + "\n")
		b.WriteString(hRow(fmt.Sprintf("Final Score: %s%d%s   Hits: %s%d%s   Missed: %s%d%s",
			FgYlw+BOLD, g.score, RST+TTBg,
			FgGrn+BOLD, g.hits, RST+TTBg,
			FgRed+BOLD, g.missed, RST+TTBg)) + "\n")
		b.WriteString(hRow(TTDim+"R=Restart │ M=Menu │ Q=Quit"+RST+TTBg) + "\n")
	} else {
		b.WriteString(hRow(TTDim+"Type letters to shoot aliens │ ESC=Menu"+RST+TTBg) + "\n")
	}
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
	stSpaceInv // Space Invaders game
)

// keyChan starts a background goroutine that reads keys and sends them on a channel.
func keyChan() <-chan keyEvent {
	ch := make(chan keyEvent, 8)
	go func() {
		buf := make([]byte, 16)
		for {
			k := readKey(buf)
			ch <- k
		}
	}()
	return ch
}

// runSpaceInvaders runs the Space Invaders game loop with its own ticker.
// Returns the action to take: "menu", "quit", or "".
func runSpaceInvaders(keys <-chan keyEvent) string {
	game := newSpaceGame()
	game.lastRender = time.Now()
	renderSpaceGame(game)

	ticker := time.NewTicker(game.tickRate)
	defer ticker.Stop()

	var mu sync.Mutex

	for {
		select {
		case <-ticker.C:
			mu.Lock()
			if !game.gameOver {
				game.update()
				renderSpaceGame(game)
				// adjust ticker if speed changed
			}
			mu.Unlock()

		case k := <-keys:
			mu.Lock()
			if k.kind == evCtrlC {
				mu.Unlock()
				return "quit"
			}
			if k.kind == evEscape {
				mu.Unlock()
				return "menu"
			}

			if game.gameOver {
				if k.kind == evChar {
					switch k.ch {
					case 'r', 'R':
						game = newSpaceGame()
						game.lastRender = time.Now()
						renderSpaceGame(game)
					case 'm', 'M':
						mu.Unlock()
						return "menu"
					case 'q', 'Q':
						mu.Unlock()
						return "quit"
					}
				}
			} else {
				if k.kind == evChar {
					ch := k.ch
					if ch >= 'A' && ch <= 'Z' {
						ch = ch - 'A' + 'a'
					}
					if game.tryShoot(ch) {
						renderSpaceGame(game)
					}
				}
			}
			mu.Unlock()
		}
	}
}

func run() error {
	fd := int(os.Stdin.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to enter raw mode: %w", err)
	}
	defer func() {
		showCur()
		term.Restore(fd, old)
	}()

	hideCur()
	keys := keyChan()
	nMenu := len(lessons) + 1 // lessons + Space Invaders
	sel := 0
	state := stMenu
	var sess *Session

	renderMenu(sel)

	for {
		k := <-keys

		// Ctrl-C always quits
		if k.kind == evCtrlC {
			cls()
			emit("Goodbye!\n")
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
				if sel < nMenu-1 {
					sel++
				}
				renderMenu(sel)
			case evEnter:
				if sel == len(lessons) {
					// Space Invaders — runs its own loop
					result := runSpaceInvaders(keys)
					switch result {
					case "quit":
						cls()
						emit("Goodbye!\n")
						return nil
					default: // "menu"
						state = stMenu
						renderMenu(sel)
					}
				} else {
					sess = newSession(&lessons[sel])
					state = stTyping
					renderTyping(sess)
				}
			case evChar:
				switch k.ch {
				case 'q', 'Q':
					cls()
					emit("Goodbye!\n")
					return nil
				default:
					if k.ch >= '1' && k.ch <= rune('0'+nMenu) {
						sel = int(k.ch - '1')
						renderMenu(sel)
					}
				}
			case evEscape:
				cls()
				emit("Goodbye!\n")
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
					emit("Goodbye!\n")
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
		fmt.Fprintf(os.Stderr, "Error: %v\r\n", err)
		os.Exit(1)
	}
}
