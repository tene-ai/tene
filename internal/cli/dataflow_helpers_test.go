package cli

import "os"

// writeFileImpl is a tiny os.WriteFile wrapper used by the dataflow test
// to seed input fixtures. The indirection is purely for test readability:
// dataflow_test.go calls writeFile(...) which delegates here.
func writeFileImpl(path, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}
