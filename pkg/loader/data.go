package loader

import _ "embed"

// EmbeddedData contains the compressed protobuf dataset
//
//go:embed combined_names.pb.gz
var EmbeddedData []byte