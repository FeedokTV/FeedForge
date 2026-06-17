package profile

import _ "embed"

//go:embed builtin/urlhaus.yaml
var BuiltinUrlhausProfile []byte

//go:embed builtin/openphish.yaml
var BuiltinOpenphishProfile []byte

//go:embed builtin/threatfox.yaml
var BuiltinThreatfoxProfile []byte

//go:embed builtin/generic_csv.yaml
var BuiltinGenericCSVProfile []byte

//go:embed builtin/generic_list.yaml
var BuiltinGenericListProfile []byte
