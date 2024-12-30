package templatex

import "errors"

var (
	ErrTemplateNotFound             = errors.New("template not found")
	ErrTemplateExecutionFailed      = errors.New("template execution failed")
	ErrTemplateParsingFailed        = errors.New("template parsing failed")
	ErrNoTemplateDirectory          = errors.New("no template directory provided")
	ErrTemplateEngineNotInitialized = errors.New("template engine not initialized")
	ErrNoTemplatesParsed            = errors.New("no templates parsed")
	ErrTemplateCloneFailed          = errors.New("failed to clone template")
)
