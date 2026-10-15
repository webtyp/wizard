package wizard

import _ "embed"

//go:embed lang.json
var dictionary []byte

// Dictionary returns wizard's translations (its lang.json, embedded), for
// lang.Load at the tool's startup.
func Dictionary() []byte { return dictionary }
