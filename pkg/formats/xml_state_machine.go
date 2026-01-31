package formats

import (
	unifiederrors "infomunge/internal/errors"
)

// XMLScannerState represents the current state of the XML scanner state machine.
type XMLScannerState int

const (
	// StateText: scanning normal text content (outside tags and strings)
	StateText XMLScannerState = iota
	// StateInString: inside a double-quoted string (in or out of tags)
	StateInString
	// StateInOpeningTag: inside an opening tag (between < and >)
	StateInOpeningTag
	// StateInClosingTag: inside a closing tag (between </ and >)
	StateInClosingTag
	// StateInComment: inside an XML comment (<!-- ... -->)
	StateInComment
	// StateInProcessingInstruction: inside a processing instruction (<? ... ?>)
	StateInProcessingInstruction
)

// XMLStateMachine validates XML structure with proper tag nesting and bracket balance.
// It uses an explicit state machine instead of nested conditionals for clarity and maintainability.
type XMLStateMachine struct {
	content string
	pos     int
	state   XMLScannerState
	stack   []string // stack of open tag names
}

// NewXMLStateMachine creates a new XML state machine for the given content.
func NewXMLStateMachine(content string) *XMLStateMachine {
	return &XMLStateMachine{
		content: content,
		pos:     0,
		state:   StateText,
		stack:   make([]string, 0),
	}
}

// Validate runs the state machine and returns an error if XML is invalid.
func (m *XMLStateMachine) Validate() error {
	for m.pos < len(m.content) {
		if err := m.processState(); err != nil {
			return err
		}
	}

	// Check for unclosed tags at end
	if len(m.stack) > 0 {
		return unifiederrors.ValidationErrorf("XML validation error: unclosed tags: %v", m.stack)
	}

	return nil
}

// processState handles the current character based on the current state.
func (m *XMLStateMachine) processState() error {
	ch := m.content[m.pos]

	switch m.state {
	case StateText:
		return m.processStateText(ch)
	case StateInString:
		return m.processStateInString(ch)
	case StateInOpeningTag:
		return m.processStateInOpeningTag(ch)
	case StateInClosingTag:
		return m.processStateInClosingTag(ch)
	case StateInComment:
		return m.processStateInComment(ch)
	case StateInProcessingInstruction:
		return m.processStateInProcessingInstruction(ch)
	default:
		return unifiederrors.ValidationErrorf("unknown state: %d", m.state)
	}
}

// processStateText handles text outside of tags.
// Transitions to other states when encountering tag markers.
func (m *XMLStateMachine) processStateText(ch byte) error {
	if ch == '<' {
		return m.handleLessThanInText()
	}
	m.pos++
	return nil
}

// handleLessThanInText processes the '<' character in text state and transitions appropriately.
func (m *XMLStateMachine) handleLessThanInText() error {
	if m.pos+1 >= len(m.content) {
		return unifiederrors.ValidationError("XML validation error: incomplete tag at end of input")
	}

	next := m.content[m.pos+1]
	m.pos += 2 // consume '<' and next character

	switch next {
	case '/':
		// Closing tag
		m.state = StateInClosingTag
		return m.parseClosingTagName()
	case '!':
		// Comment or DOCTYPE
		m.state = StateInComment
		return nil
	case '?':
		// Processing instruction
		m.state = StateInProcessingInstruction
		return nil
	default:
		// Opening tag
		m.pos-- // back up one since we'll need to process the tag name character
		m.state = StateInOpeningTag
		return m.parseOpeningTagName()
	}
}

// parseClosingTagName extracts the closing tag name and validates it.
func (m *XMLStateMachine) parseClosingTagName() error {
	tagStart := m.pos
	for m.pos < len(m.content) && m.content[m.pos] != '>' && m.content[m.pos] != ' ' && m.content[m.pos] != '\t' && m.content[m.pos] != '\n' {
		m.pos++
	}

	if m.pos >= len(m.content) {
		return unifiederrors.ValidationError("XML validation error: unclosed closing tag")
	}

	tagName := m.content[tagStart:m.pos]

	// Skip to '>'
	for m.pos < len(m.content) && m.content[m.pos] != '>' {
		m.pos++
	}

	if m.pos >= len(m.content) {
		return unifiederrors.ValidationError("XML validation error: unclosed closing tag")
	}

	// Validate tag matches opening
	if len(m.stack) == 0 || m.stack[len(m.stack)-1] != tagName {
		return unifiederrors.ValidationErrorf("XML validation error: closing tag '%s' does not match opening tag", tagName)
	}
	m.stack = m.stack[:len(m.stack)-1]
	m.pos++ // consume '>'
	m.state = StateText
	return nil
}

// parseOpeningTagName extracts the opening tag name and pushes it to the stack if not self-closing.
func (m *XMLStateMachine) parseOpeningTagName() error {
	tagStart := m.pos
	// Read tag name (stop at space, newline, /, or >)
	for m.pos < len(m.content) && m.content[m.pos] != '>' && m.content[m.pos] != ' ' && m.content[m.pos] != '\t' && m.content[m.pos] != '\n' && m.content[m.pos] != '/' {
		m.pos++
	}

	tagName := m.content[tagStart:m.pos]

	// Skip to '>' (could have attributes or self-closing marker)
	isSelfClosing := false
	for m.pos < len(m.content) && m.content[m.pos] != '>' {
		if m.content[m.pos] == '"' {
			m.pos++ // Skip opening quote
			// Skip to closing quote (handle escapes)
			for m.pos < len(m.content) && (m.content[m.pos] != '"' || (m.pos > 0 && m.content[m.pos-1] == '\\')) {
				m.pos++
			}
			if m.pos < len(m.content) {
				m.pos++ // Skip closing quote
			}
		} else if m.content[m.pos] == '/' && m.pos+1 < len(m.content) && m.content[m.pos+1] == '>' {
			isSelfClosing = true
			break
		} else {
			m.pos++
		}
	}

	if m.pos >= len(m.content) {
		return unifiederrors.ValidationError("XML validation error: unclosed opening tag")
	}

	// Add to stack if not self-closing
	if !isSelfClosing && tagName != "" {
		m.stack = append(m.stack, tagName)
	}

	m.pos++ // consume '>'
	m.state = StateText
	return nil
}

// processStateInString handles characters inside a double-quoted string.
func (m *XMLStateMachine) processStateInString(ch byte) error {
	if ch == '"' && (m.pos == 0 || m.content[m.pos-1] != '\\') {
		m.pos++
		m.state = StateText
		return nil
	}
	m.pos++
	return nil
}

// processStateInOpeningTag handles characters inside an opening tag.
// Transitions to StateInString when encountering a quote.
func (m *XMLStateMachine) processStateInOpeningTag(ch byte) error {
	if ch == '"' {
		m.state = StateInString
		m.pos++
		return nil
	}
	if ch == '>' {
		// End of opening tag - check if self-closing
		if m.pos > 0 && m.content[m.pos-1] == '/' {
			// Self-closing: don't push to stack (already handled in parseOpeningTagName)
		}
		m.state = StateText
		m.pos++
		return nil
	}
	m.pos++
	return nil
}

// processStateInClosingTag handles characters inside a closing tag.
// Transitions to StateInString when encountering a quote (edge case).
func (m *XMLStateMachine) processStateInClosingTag(ch byte) error {
	if ch == '>' {
		// End of closing tag
		m.state = StateText
		m.pos++
		return nil
	}
	m.pos++
	return nil
}

// processStateInComment handles characters inside a comment.
// Looks for --> to exit the comment.
func (m *XMLStateMachine) processStateInComment(ch byte) error {
	if ch == '-' && m.pos+2 < len(m.content) && m.content[m.pos+1] == '-' && m.content[m.pos+2] == '>' {
		m.pos += 3 // consume '-->'
		m.state = StateText
		return nil
	}
	m.pos++
	return nil
}

// processStateInProcessingInstruction handles characters inside a processing instruction.
// Looks for ?> to exit the instruction.
func (m *XMLStateMachine) processStateInProcessingInstruction(ch byte) error {
	if ch == '?' && m.pos+1 < len(m.content) && m.content[m.pos+1] == '>' {
		m.pos += 2 // consume '?>'
		m.state = StateText
		return nil
	}
	m.pos++
	return nil
}

// validateXMLBracketsWithStateMachine validates XML using the explicit state machine.
func validateXMLBracketsWithStateMachine(content string) error {
	machine := NewXMLStateMachine(content)
	return machine.Validate()
}
