package calculator

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/emil/quickcast/internal/search"
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Name() string {
	return "Calculator"
}

func (p *Provider) Refresh() error {
	return nil
}

func (p *Provider) Search(query string) ([]search.Item, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, nil
	}

	// Only attempt math evaluation if there is at least one math operator or known math function/constant
	hasMath := false
	for _, ch := range q {
		if strings.ContainsRune("+-*/%^()", ch) {
			hasMath = true
			break
		}
	}
	if !hasMath && !strings.HasPrefix(q, "sqrt") && !strings.HasPrefix(q, "sin") && !strings.HasPrefix(q, "cos") && !strings.HasPrefix(q, "tan") && !strings.HasPrefix(q, "log") {
		return nil, nil
	}

	val, err := Evaluate(q)
	if err != nil {
		return nil, nil
	}

	// Format result nicely (omit trailing .00 if whole number)
	var formatted string
	if math.Abs(val-math.Round(val)) < 1e-9 {
		formatted = fmt.Sprintf("%.0f", val)
	} else {
		formatted = strconv.FormatFloat(val, 'f', -1, 64)
	}

	markdown := fmt.Sprintf("## Calculator\n\n### `%s` = **`%s`**\n\n---\nPress `↵` to copy the result to clipboard.", q, formatted)

	item := search.Item{
		ID:          "calc:result",
		Title:       formatted,
		Subtitle:    fmt.Sprintf("= %s", q),
		Category:    "Calculator",
		Icon:        "accessories-calculator",
		Accessories: []string{"Calculator", "↵ Copy"},
		Score:       10000.0, // Pin to top when valid calculation
		Detail: &search.ItemDetail{
			Markdown: markdown,
			Metadata: map[string]string{
				"Expression": q,
				"Result":     formatted,
			},
		},
		Actions: []search.Action{
			{
				ID:       "copy_result",
				Title:    "Copy Result to Clipboard",
				Shortcut: "↵",
				Type:     "copy",
				Payload:  formatted,
			},
			{
				ID:       "copy_expr",
				Title:    "Copy Expression and Result",
				Shortcut: "Ctrl+C",
				Type:     "copy",
				Payload:  fmt.Sprintf("%s = %s", q, formatted),
			},
		},
	}

	return []search.Item{item}, nil
}

func (p *Provider) Execute(action search.Action) error {
	return nil
}

// Simple recursive-descent math expression evaluator
func Evaluate(expr string) (float64, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return 0, err
	}
	p := &parser{tokens: tokens, pos: 0}
	val, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	if p.pos < len(p.tokens) {
		return 0, fmt.Errorf("unexpected token %s", p.tokens[p.pos])
	}
	return val, nil
}

func tokenize(s string) ([]string, error) {
	var tokens []string
	runes := []rune(s)
	n := len(runes)
	i := 0

	for i < n {
		ch := runes[i]
		if unicode.IsSpace(ch) {
			i++
			continue
		}

		if unicode.IsDigit(ch) || ch == '.' {
			start := i
			for i < n && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
				i++
			}
			tokens = append(tokens, string(runes[start:i]))
			continue
		}

		if unicode.IsLetter(ch) {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i])) {
				i++
			}
			tokens = append(tokens, strings.ToLower(string(runes[start:i])))
			continue
		}

		if strings.ContainsRune("+-*/%^()", ch) {
			tokens = append(tokens, string(ch))
			i++
			continue
		}

		return nil, fmt.Errorf("unknown character: %c", ch)
	}

	return tokens, nil
}

type parser struct {
	tokens []string
	pos    int
}

func (p *parser) peek() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.pos]
}

func (p *parser) next() string {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *parser) parseExpr() (float64, error) {
	val, err := p.parseTerm()
	if err != nil {
		return 0, err
	}

	for {
		op := p.peek()
		if op == "+" || op == "-" {
			p.next()
			rhs, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			if op == "+" {
				val += rhs
			} else {
				val -= rhs
			}
		} else {
			break
		}
	}
	return val, nil
}

func (p *parser) parseTerm() (float64, error) {
	val, err := p.parsePower()
	if err != nil {
		return 0, err
	}

	for {
		op := p.peek()
		if op == "*" || op == "/" || op == "%" {
			p.next()
			rhs, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			if op == "*" {
				val *= rhs
			} else if op == "/" {
				if rhs == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				val /= rhs
			} else {
				val = math.Mod(val, rhs)
			}
		} else {
			break
		}
	}
	return val, nil
}

func (p *parser) parsePower() (float64, error) {
	val, err := p.parseFactor()
	if err != nil {
		return 0, err
	}

	for p.peek() == "^" {
		p.next()
		rhs, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		val = math.Pow(val, rhs)
	}
	return val, nil
}

func (p *parser) parseFactor() (float64, error) {
	tok := p.peek()
	if tok == "" {
		return 0, fmt.Errorf("unexpected end of expression")
	}

	if tok == "-" {
		p.next()
		val, err := p.parseFactor()
		return -val, err
	}
	if tok == "+" {
		p.next()
		return p.parseFactor()
	}

	if tok == "(" {
		p.next()
		val, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if p.peek() != ")" {
			return 0, fmt.Errorf("expected ')'")
		}
		p.next()
		return val, nil
	}

	// Functions / Constants
	switch tok {
	case "pi":
		p.next()
		return math.Pi, nil
	case "e":
		p.next()
		return math.E, nil
	case "sqrt", "sin", "cos", "tan", "abs", "ln", "log":
		p.next()
		if p.peek() != "(" {
			return 0, fmt.Errorf("expected '(' after %s", tok)
		}
		p.next()
		arg, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if p.peek() != ")" {
			return 0, fmt.Errorf("expected ')'")
		}
		p.next()

		switch tok {
		case "sqrt":
			if arg < 0 {
				return 0, fmt.Errorf("sqrt of negative number")
			}
			return math.Sqrt(arg), nil
		case "sin":
			return math.Sin(arg), nil
		case "cos":
			return math.Cos(arg), nil
		case "tan":
			return math.Tan(arg), nil
		case "abs":
			return math.Abs(arg), nil
		case "ln":
			return math.Log(arg), nil
		case "log":
			return math.Log10(arg), nil
		}
	}

	// Number literal
	p.next()
	num, err := strconv.ParseFloat(tok, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", tok)
	}
	return num, nil
}
