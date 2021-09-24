package v1alpha2

import (
	"context"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmp"

	"github.com/pivotal/kpack/pkg/apis/validate"
)

func (b *Build) SetDefaults(ctx context.Context) {
	if b.Spec.ServiceAccount == "" {
		b.Spec.ServiceAccount = "default"
	}
}

func (b *Build) Validate(ctx context.Context) *apis.FieldError {
	return b.Spec.Validate(ctx).ViaField("spec")
}

func (bs *BuildSpec) Validate(ctx context.Context) *apis.FieldError {
	return validate.ListNotEmpty(bs.Tags, "tags").
		Also(validate.Tags(bs.Tags)).
		Also(bs.Cache.Validate(ctx).ViaField("cache")).
		Also(bs.Builder.Validate(ctx).ViaField("builder")).
		Also(bs.Source.Validate(ctx).ViaField("source")).
		Also(bs.Bindings.Validate(ctx).ViaField("bindings")).
		Also(bs.LastBuild.Validate(ctx).ViaField("lastBuild")).
		Also(bs.validateImmutableFields(ctx)).
		Also(bs.validateNodeSelector(ctx))
}

func (bs *BuildSpec) validateImmutableFields(ctx context.Context) *apis.FieldError {
	if !apis.IsInUpdate(ctx) {
		return nil
	}

	original := apis.GetBaseline(ctx).(*Build)
	if diff, err := kmp.ShortDiff(&original.Spec, bs); err != nil {
		return &apis.FieldError{
			Message: "Failed to diff Build",
			Paths:   []string{"spec"},
			Details: err.Error(),
		}
	} else if diff != "" {
		return &apis.FieldError{
			Message: "Immutable fields changed (-old +new)",
			Paths:   []string{"spec"},
			Details: diff,
		}
	}
	return nil
}

func (bs *BuildSpec) validateNodeSelector(_ context.Context) *apis.FieldError {
	if len(bs.NodeSelector) == 0 {
		return nil
	}

	if _, ok := bs.NodeSelector[k8sOSLabel]; ok {
		return apis.ErrInvalidKeyName(k8sOSLabel, "nodeSelector", "os is determined automatically")
	}
	return nil

}

func (lb *LastBuild) Validate(context context.Context) *apis.FieldError {
	if lb == nil || lb.Image == "" {
		return nil
	}

	return validate.Image(lb.Image)
}

func (c *BuildCacheConfig) Validate(context context.Context) *apis.FieldError {
	if c != nil && c.Volume != nil && c.Registry != nil {
		return apis.ErrGeneric("only one type of cache can be specified", "volume", "registry")
	}
	return nil
}