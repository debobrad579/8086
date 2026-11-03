package decode_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/debobrad579/8086/internal/decode"
)

func TestDecompileToAssembly(t *testing.T) {
	testDir := "test_cases"

	testFiles, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, testFile := range testFiles {
		if testFile.IsDir() || !strings.HasSuffix(testFile.Name(), ".asm") {
			continue
		}

		t.Run(testFile.Name(), func(t *testing.T) {
			tempDir := t.TempDir()

			originalBinary := filepath.Join(tempDir, "original")
			decodedFile := filepath.Join(tempDir, "decoded.asm")
			decodedBinary := filepath.Join(tempDir, "decoded")

			cmd := exec.Command(
				"nasm",
				"-o", originalBinary,
				filepath.Join(testDir, testFile.Name()),
			)

			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("Original NASM compilation failed: %v\n%s", err, output)
			}

			originalData, err := os.ReadFile(originalBinary)
			if err != nil {
				t.Fatal(err)
			}

			decodedAsm, err := decode.DecodeInstructions(originalData)
			if err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(decodedFile, []byte(decodedAsm), 0644); err != nil {
				t.Fatal(err)
			}

			cmd = exec.Command(
				"nasm",
				"-o", decodedBinary,
				decodedFile,
			)

			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("Decoded NASM compilation failed: %v\n%s", err, output)
			}

			decodedData, err := os.ReadFile(decodedBinary)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(originalData, decodedData) {
				t.Fatalf("Decompiled assembly does not match")
			}
		})
	}
}
