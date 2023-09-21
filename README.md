# Casbin Enforce Race
Minimal reproducer for a race in `Enforce` calls, which result in invalid results.

## Local changes
Casbin is vendored locally, with the following change:
```diff
diff --git a/enforcer.go b/enforcer.go
