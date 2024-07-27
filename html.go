package main

import (
	"html"
	"strings"
)

const charSet = "UTF-8"

type htmlbuilder struct {
	strings.Builder
}

type attribute struct {
	key   string
	value string
}

var alignRight = attribute{key: "align", value: "right"}
var alignLeft = attribute{key: "align", value: "left"}
var alignCenter = attribute{key: "align", value: "center"}
var backgroundColor = attribute{"style", "background-color: #f2f2f2;"}
var backgroundWhiteColor = attribute{"style", "background-color: ##ffffff;"}
var smallCaps = attribute{"style", "font-variant: small-caps;"}

func (h *htmlbuilder) WriteOpenTag(tag string, aa ...attribute) *htmlbuilder {
	h.WriteString("<")
	h.WriteString(tag)
	for _, a := range aa {
		h.WriteString(" ")
		h.WriteString(a.key)
		h.WriteString("='")
		h.WriteString(a.value)
		h.WriteString("'")
	}
	h.WriteString(">")
	return h
}

func (h *htmlbuilder) Wrap(tag string, text string, aa ...attribute) *htmlbuilder {
	return h.WriteOpenTag(tag, aa...).Text(text).WriteCloseTag(tag)
}

func (h *htmlbuilder) WriteCloseTag(tag string) *htmlbuilder {
	h.WriteString("</")
	h.WriteString(tag)
	h.WriteString(">\n")
	return h
}

func (h *htmlbuilder) Text(t string) *htmlbuilder {
	h.WriteString(html.EscapeString(t))
	return h
}

type cycler[T any] struct {
	values []T
	offset int
}

func (c *cycler[T]) next() T {
	v := c.values[c.offset]
	c.offset = (c.offset + 1) % len(c.values)
	return v
}
