package tfcheck

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// FindBinary returns the first of terraform and tofu (OpenTofu) found in PATH.
func FindBinary() (string, error) {
	for _, name := range []string{"terraform", "tofu"} {
		_, err := exec.LookPath(name)
		if err == nil {
			return name, nil
		}
		if !errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("failed to look up '%s' in PATH: %v", name, err)
		}
	}
	return "", errors.New("neither 'terraform' nor 'tofu' found in PATH")
}

func newTerraformRunner(dir, binary, tflintConfig string) terraformRunner {
	return terraformRunner{
		dir:          dir,
		binary:       binary,
		tflintConfig: tflintConfig,
	}
}

type terraformRunner struct {
	dir          string
	binary       string
	tflintConfig string
}

func (r *terraformRunner) run(w io.Writer, command string, args ...string) error {
	cmd := exec.Command(command, args...)

	cmd.Dir = r.dir
	cmd.Stdout = w
	cmd.Stderr = w

	return cmd.Run()
}

func (r *terraformRunner) fmt(w io.Writer) error {
	err := r.run(w, r.binary, "fmt", "-list=true", "-check=true", "-recursive=false")
	if err != nil {
		return err
	}
	return nil
}

func (r *terraformRunner) init(w io.Writer) error {
	err := r.run(w, r.binary, "init", "-backend=false", "-input=false", "-get=true", "-no-color")
	if err != nil {
		return err
	}
	return nil
}

func (r *terraformRunner) validate(w io.Writer) error {
	// Validate exits cleanly even when there is a validation warning/error, so
	// we need to capture the output and look for a substring.
	var b bytes.Buffer
	w = io.MultiWriter(w, &b)

	err := r.run(w, r.binary, "validate")
	if err != nil {
		return err
	}
	if strings.Contains(b.String(), "The configuration is valid.") {
		return nil
	}
	return errors.New("validation errors")
}

func (r *terraformRunner) tflint(w io.Writer) error {
	err := r.run(w, "tflint", "--init")
	if err != nil {
		return err
	}
	args := []string{"--recursive"}
	if r.tflintConfig != "" {
		args = append(args, "--config", r.tflintConfig)
	}
	err = r.run(w, "tflint", args...)
	if err != nil {
		return err
	}
	return nil
}

// FindTerraformDirectories recursively from the given root directory.
func FindTerraformDirectories(root string) ([]string, error) {
	dirs := make(map[string]struct{})
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == ".terraform" {
			return filepath.SkipDir
		}
		if !info.IsDir() && filepath.Ext(path) == ".tf" {
			dirs[filepath.Dir(path)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var output []string
	for dir := range dirs {
		output = append(output, dir)
	}

	sort.Strings(output)
	return output, nil
}
