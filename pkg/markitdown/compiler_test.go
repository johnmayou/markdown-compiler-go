package markitdown

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func TestCompile(t *testing.T) {
	cases, err := filepath.Glob("testdata/*.text")
	require.NoError(t, err)

	for _, casepath := range cases {
		t.Run(casepath, func(t *testing.T) {
			md, err := os.ReadFile(casepath)
			require.NoError(t, err)

			actual, err := Compile(string(md))
			require.NoError(t, err)

			actual, err = prettifyHTML(actual)
			require.NoError(t, err)

			golden := replaceExt(casepath, "html")
			if *update {
				require.NoError(t, os.WriteFile(golden, []byte(actual), 0644))
			}

			expected, err := os.ReadFile(golden)
			require.NoError(t, err)
			require.Equal(t, string(expected), actual)
		})
	}
}

func prettifyHTML(html string) (string, error) {
	cmd := exec.Command("prettier", "--parser", "html")
	cmd.Stdin = strings.NewReader(html)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("prettier format: %w", err)
	}
	return string(out), nil
}

func replaceExt(path, new string) string {
	return strings.TrimSuffix(path, filepath.Ext(path)) + "." + new
}
