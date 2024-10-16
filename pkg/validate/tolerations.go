// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"

	build "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
)

// TolerationsRef contains all required fields
// to validate tolerations
type TolerationsRef struct {
	Build *build.Build // build instance for analysis
}

func NewTolerations(build *build.Build) *TolerationsRef {
	return &TolerationsRef{build}
}

// ValidatePath implements BuildPath interface and validates
// that tolerations key/operator/value are valid
func (b *TolerationsRef) ValidatePath(_ context.Context) error {
	for _, toleration := range b.Build.Spec.Tolerations {
		if errs := validation.IsValidLabelValue(toleration.Key); errs != nil {
			b.Build.Status.Reason = ptr.To(build.TolerationNotValid)
			b.Build.Status.Message = ptr.To(strings.Join(errs, ", "))
		}
	}

	return nil
}
