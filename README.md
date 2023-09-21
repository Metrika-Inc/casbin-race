# Casbin Enforce Race
Minimal reproducer for a race in `Enforce` calls, which result in invalid results.

## Reproduction
This will reproduce the failure most of the time, but there is a theoretical chance of it succeeding.
```
$ make test
go test ./...
panic: result failure: 1

goroutine 28 [running]:
casbin_race.TestRaceFail.func1(0x0?)
        /home/user/repos/casbin-race/race_test.go:69 +0x134
created by casbin_race.TestRaceFail in goroutine 6
        /home/user/repos/casbin-race/race_test.go:62 +0x825
FAIL    casbin_race     5.251s
FAIL
make: *** [Makefile:2: test] Error 1
```

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
The bug _seems_ to be caused by concurrent modification of the roles in `RoleManagerImpl`. During calls to `HasLink` roles are created if they do not exist, and removed after the function returns. This means that if two calls
to `HasLink` are made concurrently, the second may see the existing (temporary) role, and then it may be removed before
it has been inspected. For example:
* Enforce A: Temp role does not exist, is added to map && `defer rm.removeRole()` called
* Enforce A: `hasLinkHelper` iterates over roles via `sync.Map.Range()` recursively
* Enforce B: Temp role exists, fetch it.
* Enforce A: Exit function, `removeRole()` called.
* Enforce B: Temp role has no links it was removed, cannot resolve.

**Note:** see [sync.Map.Range documentation](https://pkg.go.dev/sync#Map.Range), as it states, it does not ensure a consistent snapshot.
```
func (m *Map) Range(f func(key, value any) bool)

Range calls f sequentially for each key and value present in the map. If f returns false, range stops the iteration.

Range does not necessarily correspond to any consistent snapshot of the Map's contents: no key will be visited more than once, but if the value for any key is stored or deleted concurrently (including by f), Range may reflect any mapping for that key from any point during the Range call. Range does not block other methods on the receiver; even f itself may call any method on m.

Range may be O(N) with the number of elements in the map even if f returns false after a constant number of calls.


```

Note that the issue will still occur with the `SyncedEnforcer` because multiple calls to `Enforce` are only protected by
a non-exclusive read lock.
