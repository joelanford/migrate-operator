package apis

import (
	originalv1alpha1 "github.com/joelanford/migrate-operator/pkg/apis/original/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, originalv1alpha1.SchemeBuilder.AddToScheme)
}
