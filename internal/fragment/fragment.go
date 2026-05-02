package fragment

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"dario.cat/mergo"
	"go.yaml.in/yaml/v3"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

func New(dir string) *Fragment {
	return &Fragment{dir: dir}
}

func (f *Fragment) Read() (*Metadata, error) {
	absPath, err := filepath.Abs(f.dir)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("stat fragment path %q: %w", absPath, err)
	}

	meta := &Metadata{
		Dir:   absPath,
		Files: make([]string, 0, 64),
	}

	if !info.IsDir() {
		meta.Dir = filepath.Dir(absPath)
	}

	err = filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		switch filepath.Ext(path) {
		case ".yaml", ".yml":
		default:
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return fmt.Errorf("stat fragment file %q: %w", path, infoErr)
		}
		if info.Size() > maxFileSize {
			return fmt.Errorf("fragment file %q exceeds size limit of %d bytes", path, maxFileSize)
		}

		if err := f.readFile(path); err != nil {
			return err
		}
		meta.Files = append(meta.Files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Namespace defaults to the directory basename when not declared in YAML.
	// It is both the fragment's identity (used in logs/errors) and the prefix
	// applied to all receiver names and route receiver references during assembly.
	if f.Namespace == "" {
		f.Namespace = filepath.Base(meta.Dir)
	}

	return meta, f.validate()
}

func (f *Fragment) validate() error {
	definedReceivers := make(map[string]struct{}, len(f.Receivers))

	var errs []error
	for i, r := range f.Receivers {
		name := r.Name
		if name == "" {
			errs = append(errs, fmt.Errorf("fragment %q receiver[%d] missing or non-string name field", f.Namespace, i))
			continue
		}
		definedReceivers[name] = struct{}{}
	}

	for _, route := range f.Routes {
		if _, ok := definedReceivers[route.Receiver]; !ok {
			errs = append(errs, fmt.Errorf("fragment %q uses undefined receiver %q", f.Namespace, route.Receiver))
		}
	}

	return errors.Join(errs...)
}

func (f *Fragment) readFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read fragment file %q: %w", file, err)
	}

	if isTestFile(file) {
		var doc TestDoc
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("parse yaml in %q: %w", file, err)
		}
		for _, t := range doc.Tests {
			t.Type = "unit"
		}
		f.Tests = append(f.Tests, doc.Tests...)
		return nil
	}

	var frag Fragment
	if err := yaml.Unmarshal(data, &frag); err != nil {
		return fmt.Errorf("parse yaml in %q: %w", file, err)
	}

	return f.merge(&frag)
}

func isTestFile(file string) bool {
	if strings.Contains(filepath.Base(file), "-tests") {
		return true
	}
	dir := filepath.ToSlash(filepath.Dir(file))
	return slices.Contains(strings.Split(dir, "/"), "tests")
}

func (f *Fragment) merge(src *Fragment) error {
	if hasConflictString(f.Namespace, src.Namespace) {
		return fmt.Errorf("found conflicting fragment namespace: src=%q, dst=%q", src.Namespace, f.Namespace)
	}
	if hasConflictGroup(f.Group, src.Group) {
		return fmt.Errorf("found conflicting fragment group: src=%#v, dst=%#v", src.Group, f.Group)
	}

	return mergo.Merge(f, src, mergo.WithAppendSlice)
}

func hasConflictString(src, dst string) bool {
	return dst != "" && src != "" && src != dst
}

func hasConflictGroup(src, dst *FragmentGroup) bool {
	if isEmptyGroup(src) || isEmptyGroup(dst) {
		return false
	}
	return src.Receiver != dst.Receiver || !maps.Equal(src.Match, dst.Match)
}

func isEmptyGroup(group *FragmentGroup) bool {
	return group == nil || group.Receiver == "" && len(group.Match) == 0
}

func Load(dir string) (*LoadResult, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read fragments dir %q: %w", dir, err)
	}

	type resultItem struct {
		aug AugmentedFragment
		err error
	}

	results := make(chan resultItem, len(entries))
	var wg sync.WaitGroup

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		wg.Add(1)
		go func(entry fs.DirEntry) {
			defer wg.Done()
			path := filepath.Join(dir, entry.Name())
			frag := New(path)
			meta, err := frag.Read()
			if err != nil {
				results <- resultItem{err: err}
				return
			}
			results <- resultItem{aug: AugmentedFragment{
				Fragment: frag,
				Metadata: meta,
				Tests:    frag.Tests,
			}}
		}(entry)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	finalResult := &LoadResult{
		Fragments: make([]AugmentedFragment, 0, len(entries)),
	}

	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		finalResult.Fragments = append(finalResult.Fragments, res.aug)
	}

	return finalResult, nil
}
