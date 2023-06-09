/*
Copyright 2023 Michael Bridgen.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"fmt"
	"strings"
)

type token struct {
	text string
	ref  string
}

// For tests and troubleshooting
func (t token) String() string {
	if t.ref != "" {
		return fmt.Sprintf("${%s}", t.ref)
	}
	return fmt.Sprintf("%q", t.text)
}

// parseInterpolation takes a string value and tries to parse it as an
// interpolated string. If it cannot parse it, an error is
// returned. If it can be parsed, the list of tokens is returned
func parseInterpolation(s string) ([]token, error) {
	var parts []token
	i := 0
	var sb strings.Builder

	const (
		stateText = iota
		stateDollar
		stateRef
	)
	state := stateText

	for i < len(s) {
		switch state {
		case stateText:
			if s[i] == '$' {
				state = stateDollar
			} else {
				sb.WriteByte(s[i])
			}
		case stateDollar:
			if s[i] == '{' { // var ref ${
				if sb.Len() > 0 {
					parts = append(parts, token{text: sb.String()})
					sb.Reset()
				}
				state = stateRef
			} else if s[i] == '$' { // escaped '$' e.g., $$
				state = stateText
				sb.WriteByte('$')
			} else { // '$' in regular text, e.g., $a
				state = stateText
				sb.WriteByte('$')
				sb.WriteByte(s[i])
			}
		case stateRef:
			if s[i] == '}' {
				parts = append(parts, token{ref: sb.String()})
				sb.Reset()
				state = stateText
			} else {
				sb.WriteByte(s[i])
			}
		}
		i++
	}

	// the only problem is if you don't close a ref
	if state == stateRef {
		return nil, fmt.Errorf("malformed string value %q", s)
	}
	if sb.Len() > 0 {
		parts = append(parts, token{text: sb.String()})
	}
	return parts, nil
}

func interpolateString(e *env, template string) (string, error) {
	parts, err := parseInterpolation(template)
	if err != nil {
		return "", err
	}

	// shortcut: any values without refs will just be [token{text: <s>}]
	if len(parts) == 1 && parts[0].ref == "" {
		return parts[0].text, nil
	}

	var sb strings.Builder
	for i := range parts {
		if ref := parts[i].ref; ref != "" {
			if v, ok := e.lookup(ref); ok {
				sb.WriteString(v)
			} else {
				return "", fmt.Errorf("unknown ref %q", ref)
			}
		} else {
			sb.WriteString(parts[i].text)
		}
	}
	return sb.String(), nil
}
