package apis

import (
	migratedv1 "github.com/joelanford/migrate-operator/pkg/apis/migrated/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, migratedv1.SchemeBuilder.AddToScheme)
}
