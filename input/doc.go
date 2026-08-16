// Package input is the shared text-entry substrate: thin wrappers over
// bubbles/textinput and textarea ([Line], [Area]) plus a textarea factory with
// overlay-friendly defaults ([NewTextArea]) and a multi-entry title+body
// editor ([Editor]), and the external-editor hop ([ExternalEditorCmd]).
// Routing every prompt through one package keeps cursor movement, word-wise
// editing, and bracketed paste behaving identically everywhere.
package input
