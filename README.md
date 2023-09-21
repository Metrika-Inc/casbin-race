# Casbin Enforce Race
Minimal reproducer for a race in `Enforce` calls, which result in invalid results.

## Local changes
Casbin is vendored locally, with the following change:

```diff
diff --git a/vendor/github.com/casbin/casbin/v2/util/builtin_operators.go b/vendor/github.com/casbin/casbin/v2/util/builtin_operators.go
index 2206d07..8e896cf 100644
--- a/vendor/github.com/casbin/casbin/v2/util/builtin_operators.go
+++ b/vendor/github.com/casbin/casbin/v2/util/builtin_operators.go
@@ -407,6 +407,7 @@ func GenerateGFunction(rm rbac.RoleManager) govaluate.ExpressionFunction {
 
                // ...and see if we've already calculated this.
                v, found := memorized.Load(key)
+               found = false
                if found {
                        return v, nil
                }
```

This change will result in computation not being cached, which makes the bug significantly easier to reproduce. That
said, the bug is still present (and was discovered) without this change.

## The bug (or a theory)
The bug _seems_ to be caused by concurrent modification of the roles in `RoleManagerImpl`. During calls to `HasLink` (
and the recursive helper), roles are created if they do not exist, and removed afterward. This means that if two calls
to `HasLink` are made concurrently, the second may see the existing (temporary) role, and then it may be removed before
it has been inspected. For example:
* Enforce A: Temp role added
* Enforce A: Inspect temp role (recursively)
* Enforce B: Fetch the same temp role, no need to create, it exists
* Enforce A: Delete temp role
* Enforce B: Temp role has no links it was removed, cannot resolve.

Note that the issue will still occur with the `SyncedEnforcer` because multiple calls to `Enforce` are only protected by
a non-exclusive read lock.
