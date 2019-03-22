# migrate-operator

This repo contains branches that demonstrate how to migrate Operator SDK APIs and Controllers

The assumption in this repo is that the old and new APIs are functionally identical (i.e. the
APIs have the exact same fields, and the Reconciler performs the exact same reconciliation).

## Repo branches

- [master](https://github.com/joelanford/migrate-operator/tree/master) - An initial project with API `original.com/v1alpha1`

- [migrate-version](https://github.com/joelanford/migrate-operator/tree/migrate-version) - Builds on `master` by:
  - Renaming `v1alpha1` to `v1` in directory, files, and code
  - Ensuring that both `v1alpha1` and `v1` are included in the CRD versions.

- [migrate-group-and-version](https://github.com/joelanford/migrate-operator/tree/migrate-group-and-version) - Builds on `master` by:
  - Renaming `original.com/v1alpha1` to `migrated.com/v1` in directory, files, and code.
  - Re-adding `original.com/v1alpha1` API, whose type declarations use types from `migrated.com/v1`
  - Updating the controller to watch both APIs and to use `unstructured.Unstructured` to generically fetch and convert the reconciled resources.
