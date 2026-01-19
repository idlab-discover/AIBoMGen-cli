package metadata

import "fmt"

type applyInput struct {
	Value any
	Force bool
}

// ApplyFromSources applies the first available source value using spec.Apply.
func ApplyFromSources(spec FieldSpec, src Source, tgt Target) {
	if spec.Apply == nil || len(spec.Sources) == 0 {
		return
	}
	for _, get := range spec.Sources {
		if get == nil {
			continue
		}
		value, ok := get(src)
		if !ok {
			continue
		}
		_ = spec.Apply(tgt, applyInput{Value: value, Force: false})
		return
	}
}

// ApplyUserValue parses and applies a user-provided value using spec.Parse and spec.Apply.
func ApplyUserValue(spec FieldSpec, value string, tgt Target) error {
	if spec.Parse == nil || spec.Apply == nil {
		return fmt.Errorf("spec missing Parse/Apply for %s", spec.Key)
	}
	parsed, err := spec.Parse(value)
	if err != nil {
		return err
	}
	return spec.Apply(tgt, applyInput{Value: parsed, Force: true})
}

// ApplyDatasetFromSources applies the first available dataset source value.
func ApplyDatasetFromSources(spec DatasetFieldSpec, src DatasetSource, tgt DatasetTarget) {
	if spec.Apply == nil || len(spec.Sources) == 0 {
		return
	}
	for _, get := range spec.Sources {
		if get == nil {
			continue
		}
		value, ok := get(src)
		if !ok {
			continue
		}
		_ = spec.Apply(tgt, applyInput{Value: value, Force: false})
		return
	}
}

// ApplyDatasetUserValue parses and applies a dataset user value.
func ApplyDatasetUserValue(spec DatasetFieldSpec, value string, tgt DatasetTarget) error {
	if spec.Parse == nil || spec.Apply == nil {
		return fmt.Errorf("spec missing Parse/Apply for %s", spec.Key)
	}
	parsed, err := spec.Parse(value)
	if err != nil {
		return err
	}
	return spec.Apply(tgt, applyInput{Value: parsed, Force: true})
}
