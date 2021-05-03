# Dependencies between Reconcile and Delete Tasks

The [reconcile operation](../pkg/virtualgarden/operation_reconcile.go) is divided into tasks. The following graph shows 
the dependencies between these tasks.  The [delete operation](../pkg/virtualgarden/operation_delete.go) is divided into 
corresponding tasks with reversed dependencies.

![dependencies between reconcile tasks](img/task-dependencies.png)

