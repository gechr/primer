// Package form is the modal text-input controller for overlay forms: an
// ordered set of single- and multi-line fields with one focus ring, a
// dirty-discard guard (esc on edited content asks yes/no instead of dropping
// it), a submit/cancel/editor event contract with a consumed-key signal, and
// a pluggable autocomplete seam (a trigger rune such as @ starts a query;
// suggestions are fetched asynchronously as commands). The package is
// domain-free - it knows nothing about what the fields mean - so any overlay
// that collects text can ride it. It is state structs plus render helpers,
// not a tea.Model: the owner routes messages in and places the rendered
// content. It is deliberately not input.Editor: that is a full tea.Model over
// a fixed title+body pair, where this package mixes one-line and multiline
// fields, gates required ones, and exposes the completion seam.
package form
