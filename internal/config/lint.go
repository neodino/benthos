package config

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"

	"github.com/benthosdev/benthos/v4/internal/docs"
	"github.com/benthosdev/benthos/v4/internal/filepath/ifs"
)

// ReadYAMLFileLinted will attempt to read a configuration file path into a
// structure. Returns an array of lint messages or an error.
func ReadYAMLFileLinted(fs ifs.FS, path string, skipEnvVarCheck bool, lConf docs.LintConfig, config *Type) ([]docs.Lint, error) {
	configBytes, lints, _, err := ReadFileEnvSwap(fs, path, os.LookupEnv)
	if err != nil {
		return nil, err
	}

	if skipEnvVarCheck {
		var newLints []docs.Lint
		for _, l := range lints {
			if l.Type != docs.LintMissingEnvVar {
				newLints = append(newLints, l)
			}
		}
		lints = newLints
	}

	cNode, err := docs.UnmarshalYAML(configBytes)
	if err != nil {
		return nil, err
	}
	if err = config.FromAny(lConf.DocsProvider, cNode); err != nil {
		return nil, err
	}

	if !bytes.HasPrefix(configBytes, []byte("# BENTHOS LINT DISABLE")) {
		lints = append(lints, Spec().LintYAML(docs.NewLintContext(lConf), cNode)...)
	}
	return lints, nil
}

// LintYAMLBytes attempts to report errors within a user config. Returns a slice of
// lint results.
func LintYAMLBytes(lintConf docs.LintConfig, rawBytes []byte) ([]docs.Lint, error) {
	if bytes.HasPrefix(rawBytes, []byte("# BENTHOS LINT DISABLE")) {
		return nil, nil
	}

	var rawNode yaml.Node
	if err := yaml.Unmarshal(rawBytes, &rawNode); err != nil {
		return nil, err
	}

	return Spec().LintYAML(docs.NewLintContext(lintConf), &rawNode), nil
}

// ReadFileEnvSwap reads a file and replaces any environment variable
// interpolations before returning the contents. Linting errors are returned if
// the file has an unexpected higher level format, such as invalid utf-8
// encoding.
//
// An modTime timestamp is returned if the modtime of the file is available.
func ReadFileEnvSwap(store ifs.FS, path string, lookupEnvFn func(name string) (string, bool)) (configBytes []byte, lints []docs.Lint, modTime time.Time, err error) {
	var configFile fs.File
	if configFile, err = store.Open(path); err != nil {
		return
	}

	if info, ierr := configFile.Stat(); ierr == nil {
		modTime = info.ModTime()
	}

	if configBytes, err = io.ReadAll(configFile); err != nil {
		return
	}

	if !utf8.Valid(configBytes) {
		lints = append(lints, docs.NewLintError(
			1, docs.LintFailedRead,
			errors.New("detected invalid utf-8 encoding in config, this may result in interpolation functions not working as expected"),
		))
	}

	if configBytes, err = ReplaceEnvVariables(configBytes, lookupEnvFn); err != nil {
		var errEnvMissing *ErrMissingEnvVars
		if errors.As(err, &errEnvMissing) {
			configBytes = errEnvMissing.BestAttempt
			lints = append(lints, docs.NewLintError(1, docs.LintMissingEnvVar, err))
			err = nil
		} else {
			return
		}
	}
	return
}
