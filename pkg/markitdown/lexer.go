package markitdown

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type token interface {
	isToken()
}

type headerToken struct {
	size int
}

type textToken struct {
	text   string
	bold   bool
	italic bool
}

type listItemToken struct {
	indent  int
	ordered bool
	digit   int
}

type codeBlockToken struct {
	lang string
	code string
}

type codeInlineToken struct {
	lang string
	code string
}

type blockQuoteToken struct {
	indent int
}

type imageToken struct {
	alt string
	src string
}

type linkToken struct {
	text string
	href string
}

type horizontalRuleToken struct{}
type newLineToken struct{}

func (t *headerToken) isToken()         {}
func (t *textToken) isToken()           {}
func (t *listItemToken) isToken()       {}
func (t *codeBlockToken) isToken()      {}
func (t *codeInlineToken) isToken()     {}
func (t *blockQuoteToken) isToken()     {}
func (t *imageToken) isToken()          {}
func (t *linkToken) isToken()           {}
func (t *horizontalRuleToken) isToken() {}
func (t *newLineToken) isToken()        {}

const listIndentSize = 2

type lexer struct {
	md  string
	tks []token
}

// newLexer returns a lexer initialized with the given markdown string.
func newLexer(md string) *lexer {
	return &lexer{md: md, tks: make([]token, 0)}
}

// tokenize processes l.md into tokens, always returning a slice ending with a newLineToken.
func (l *lexer) tokenize() ([]token, error) {
	tokenizers := []func() (bool, error){
		l.tryTokenizeHeader,
		l.tryTokenizeCodeBlock,
		l.tryTokenizeBlockQuote,
		l.tryTokenizeHorizontalRule,
		l.tryTokenizeList,
		l.tryTokenizeHeaderAlt,
		l.tryTokenizeNewLine,
	}

	for len(l.md) > 0 {
		matched := false
		for _, try := range tokenizers {
			ok, err := try()
			if err != nil {
				return nil, err
			}
			if ok {
				matched = true
				break
			}
		}
		if !matched {
			l.tokenizeCurrentLine()
		}
	}

	if len(l.tks) > 0 {
		if _, ok := l.tks[len(l.tks)-1].(*newLineToken); !ok {
			l.tks = append(l.tks, &newLineToken{})
		}
	}

	return l.tks, nil
}

var headerRegex = regexp.MustCompile(`\A(######|#####|####|###|##|#) `)

// tryTokenizeHeader matches ATX-style headers (# through ######) at the start of l.md.
func (l *lexer) tryTokenizeHeader() (bool, error) {
	match := headerRegex.FindStringSubmatch(l.md)
	if len(match) == 0 {
		return false, nil
	}

	hsize := len(match[1])
	l.tks = append(l.tks, &headerToken{size: hsize})
	l.md = l.md[hsize+1:]
	l.tokenizeCurrentLine()
	l.tks = append(l.tks, &horizontalRuleToken{})
	l.tks = append(l.tks, &newLineToken{})

	return true, nil
}

var codeBlockRegex = regexp.MustCompile("\\A```(.*?)\\s*\n")

// tryTokenizeCodeBlock matches a fenced code block, returning false if no closing ``` is found.
func (l *lexer) tryTokenizeCodeBlock() (bool, error) {
	match := codeBlockRegex.FindStringSubmatch(l.md)
	if len(match) == 0 {
		return false, nil
	}

	codeStart := 0
	for {
		codeStart++
		if codeStart >= len(l.md) {
			return false, nil // no ending to code block
		}
		if l.md[codeStart] == '\n' {
			codeStart++
			break
		}
	}

	codeEnd := codeStart
	for {
		if codeEnd+2 >= len(l.md) {
			return false, nil // no ending to code block
		}
		if l.md[codeEnd] == '`' && l.md[codeEnd+1] == '`' {
			codeEnd -= 2 // go backwards through `, \n
			break
		}
		codeEnd++
	}

	// Make sure we haven't altered `l.md` up until this point since we need to ensure there
	// is an ending block. If there is no ending block, we would have returned False somewhere
	// above. In that case, we should try tokenizing a different token with the original `l.md`.

	code := l.md[codeStart : codeEnd+2]
	l.md = l.md[codeEnd+5:] // move to after the ``` and \n
	l.tks = append(l.tks, &codeBlockToken{lang: match[1], code: code})
	l.tks = append(l.tks, &newLineToken{})

	return true, nil
}

var blockQuoteRegex = regexp.MustCompile(`\A(>(?: >)* ?)`)

// tryTokenizeBlockQuote matches one or more leading > characters at the start of l.md.
func (l *lexer) tryTokenizeBlockQuote() (bool, error) {
	match := blockQuoteRegex.FindStringSubmatch(l.md)
	if len(match) == 0 {
		return false, nil
	}

	indent := 0
	for _, ch := range match[1] {
		if ch == '>' {
			indent++
		}
	}

	l.tks = append(l.tks, &blockQuoteToken{indent: indent})
	l.md = l.md[len(match[1]):]
	l.tokenizeCurrentLine()

	return true, nil
}

var horizontalRuleRegex = regexp.MustCompile(`(?m)\A(?:\*{3,}\**|-{3,}-*)\s*$`)

// tryTokenizeHorizontalRule matches a line of three or more * or - characters.
func (l *lexer) tryTokenizeHorizontalRule() (bool, error) {
	match := horizontalRuleRegex.FindStringSubmatch(l.md)
	if len(match) == 0 {
		return false, nil
	}

	l.tks = append(l.tks, &horizontalRuleToken{})
	l.tks = append(l.tks, &newLineToken{})
	l.deleteCurrentLine()

	return true, nil
}

var listRegex = regexp.MustCompile(`\A *(?:([0-9]\.)|(\*|-)) `)

// tryTokenizeList matches ordered (1.) and unordered (- or *) list items.
func (l *lexer) tryTokenizeList() (bool, error) {
	match := listRegex.FindStringSubmatch(l.md)
	if len(match) == 0 {
		return false, nil
	}

	spaces := 0
	for l.md[spaces] == ' ' {
		spaces++
	}

	if l.md[spaces] == '*' || l.md[spaces] == '-' { // un-ordered
		l.tks = append(l.tks, &listItemToken{indent: spaces / listIndentSize, ordered: false, digit: -1})
		l.md = l.md[spaces+2:] // 2 = */- + space
	} else {
		digit, err := strconv.Atoi(string(l.md[spaces]))
		if err != nil {
			return false, fmt.Errorf("parsing ordered digit: %w", err)
		}
		l.tks = append(l.tks, &listItemToken{indent: spaces / listIndentSize, ordered: false, digit: digit})
		l.md = l.md[spaces+3:] // 3 = digit + period + space
	}

	l.tokenizeCurrentLine()

	return true, nil
}

var headerAltRegex = regexp.MustCompile(`\A.+\n(=+|-+) *`)

// tryTokenizeHeaderAlt matches setext-style headers underlined with === (h1) or --- (h2).
func (l *lexer) tryTokenizeHeaderAlt() (bool, error) {
	match := headerAltRegex.FindStringSubmatch(l.md)
	if len(match) == 0 {
		return false, nil
	}

	// Search next line for header size (=== for h1, --- for h2)
	pointer := 0
	for l.md[pointer] != '\n' {
		pointer++
	}

	hsizeCh := l.md[pointer+1] // go to beginning of next line
	switch hsizeCh {
	case '=':
		l.tks = append(l.tks, &headerToken{size: 1})
	case '-':
		l.tks = append(l.tks, &headerToken{size: 2})
	default:
		return false, fmt.Errorf("unexpected char for header alt: %c", hsizeCh)
	}

	l.tokenizeCurrentLine()
	l.deleteCurrentLine() // ---/=== line
	l.tks = append(l.tks, &horizontalRuleToken{})
	l.tks = append(l.tks, &newLineToken{})

	return true, nil
}

var newLineRegex = regexp.MustCompile(`\A\n`)

// tryTokenizeNewLine matches a bare newline at the start of l.md.
func (l *lexer) tryTokenizeNewLine() (bool, error) {
	match := newLineRegex.FindStringSubmatch(l.md)
	if len(match) == 0 {
		return false, nil
	}

	l.tks = append(l.tks, &newLineToken{})
	l.md = l.md[1:]

	return true, nil
}

var (
	textBoldItalicRegex = regexp.MustCompile(`\A(\*{3}[^\*]+?\*{3}|_{3}[^_]+?_{3})`)
	textBoldRegex       = regexp.MustCompile(`\A(\*{2}[^\*]+?\*{2}|_{2}[^_]+?_{2})`)
	textItalicRegex     = regexp.MustCompile(`\A(\*[^\*]+?\*|_[^_]+?_)`)
	textImageRegex      = regexp.MustCompile(`\A!\[(.*)\]\((.*)\)`)
	textLinkRegex       = regexp.MustCompile(`\A\[(.*?)\]\((.*?)\)`)
	textCodeRegex       = regexp.MustCompile("\\A`(.+?)`([a-z]*)")
)

// tokenizeCurrentLine tokenizes inline elements (bold, italic, links, images, inline code) within the current line.
func (l *lexer) tokenizeCurrentLine() {
	// Nothing to parse.
	if len(l.md) == 0 {
		return
	}

	// Already at the end of the line.
	if l.md[0] == '\n' {
		l.tks = append(l.tks, &newLineToken{})
		l.md = l.md[1:]
		return
	}

	// Find current line.
	lineEnd := 0
	for {
		lineEnd++
		if lineEnd == len(l.md) { // EOF
			lineEnd--
			break
		}
		if l.md[lineEnd] == '\n' {
			break
		}
	}
	line := l.md[:lineEnd+1]
	l.md = l.md[len(line):]

	// Keep track of current substring.
	var currStr []byte
	currPush := func() {
		if len(currStr) > 0 {
			l.tks = append(l.tks, &textToken{text: string(currStr), bold: false, italic: false})
			currStr = currStr[:0]
		}
	}

	// Helper for removing bold/italic chars (`*`, `_`)
	emphasisReplacer := strings.NewReplacer("*", "", "_", "")

	for len(line) > 0 {
		var match []string

		// == Bold & Italic ==
		match = textBoldItalicRegex.FindStringSubmatch(line)
		if len(match) > 0 {
			currPush()

			l.tks = append(l.tks, &textToken{text: emphasisReplacer.Replace(match[1]), bold: true, italic: true})
			line = line[len(match[1]):]

			continue
		}

		// == Bold ==
		match = textBoldRegex.FindStringSubmatch(line)
		if len(match) > 0 {
			currPush()

			l.tks = append(l.tks, &textToken{text: emphasisReplacer.Replace(match[1]), bold: true, italic: false})
			line = line[len(match[1]):]

			continue
		}

		// == Italic ==
		match = textItalicRegex.FindStringSubmatch(line)
		if len(match) > 0 {
			currPush()

			l.tks = append(l.tks, &textToken{text: emphasisReplacer.Replace(match[1]), bold: false, italic: true})
			line = line[len(match[1]):]

			continue
		}

		// == Image ==
		match = textImageRegex.FindStringSubmatch(line)
		if len(match) > 0 {
			currPush()

			l.tks = append(l.tks, &imageToken{alt: match[1], src: match[2]})
			line = line[len(match[1])+len(match[2])+5:] // 5 = ![]()

			continue
		}

		// == Link ==
		match = textLinkRegex.FindStringSubmatch(line)
		if len(match) > 0 {
			currPush()

			l.tks = append(l.tks, &linkToken{text: match[1], href: match[2]})
			line = line[len(match[1])+len(match[2])+4:] // 4 = []()

			continue
		}

		// == Code ==
		match = textCodeRegex.FindStringSubmatch(line)
		if len(match) > 0 {
			currPush()

			code := match[1]
			lang := match[2]
			l.tks = append(l.tks, &codeInlineToken{lang: lang, code: code})
			line = line[len(code)+len(lang)+2:] // 2 = ``

			continue
		}

		// == New Line ==
		if line[0] == '\n' {
			currPush()

			l.tks = append(l.tks, &newLineToken{})
			break
		}

		// == Default ==
		currStr = append(currStr, line[0])
		line = line[1:]
	}

	currPush()
}

// deleteCurrentLine removes the current line from l.md, including its trailing newline.
func (l *lexer) deleteCurrentLine() {
	pointer := 0
	for {
		pointer++
		if pointer == len(l.md) { // EOF
			l.md = ""
			break
		}
		if l.md[pointer] == '\n' {
			l.md = l.md[pointer+1:]
			break
		}
	}
}
