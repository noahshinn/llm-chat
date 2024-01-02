package stringsx

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

func AddSyntaxHighlightingToCode(code string, language string) (string, error) {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = formatter.Format(buf, style, iterator)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func AddPossibleSyntaxHighlighting(text string) (string, error) {
	regex := regexp.MustCompile("```(\\w+)([\\s\\S]*?)(```|$)")
	matches := regex.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return text, nil
	}
	var err error
	for _, match := range matches {
		language := text[match[2]:match[3]]
		codeBlock := text[match[0]:match[1]]
		code := strings.TrimPrefix(codeBlock, "```"+language)
		code = strings.TrimSuffix(code, "```")
		code = strings.TrimSpace(code)
		highlightedCode, highlightErr := AddSyntaxHighlightingToCode(code, language)
		if highlightErr != nil {
			err = highlightErr
			continue
		}
		text = strings.Replace(text, codeBlock, highlightedCode, 1)
	}

	return text, err
}
