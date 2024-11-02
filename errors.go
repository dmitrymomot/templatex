package templatex

import "errors"

var (
	ErrTemplateNotFound        = errors.New("template not found")
	ErrTemplateExecutionFailed = errors.New("template execution failed")
	ErrTemplateParsingFailed   = errors.New("template parsing failed")
	ErrNoTemplateDirectory     = errors.New("no template patterns provided")
)
