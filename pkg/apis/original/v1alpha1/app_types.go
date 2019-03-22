package v1alpha1

import (
	migratedv1 "github.com/joelanford/migrate-operator/pkg/apis/migrated/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// App is the Schema for the apps API
// +k8s:openapi-gen=true
type App migratedv1.App

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppList contains a list of App
type AppList migratedv1.AppList

func init() {
	SchemeBuilder.Register(&App{}, &AppList{})
}
