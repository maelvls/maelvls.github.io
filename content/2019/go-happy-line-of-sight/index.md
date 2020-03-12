---
title: "Go Happy Path: the Unindented Line of Sight"
tags: [go]
date: 2019-11-23
description: |
    Readability is a properly we all love about Go. In other languages, it
    might be fine to have a lot of nested if statements; in Go, it is a good
    practice to keep away from overly-nested logic.
url: /go-happy-line-of-sight
---

While perusing how other Kubernetes developers are implementing their own
reconciliation loop, I came across [an interesting piece of
code](https://github.com/kubeflow/katib/blob/40f55b41c/pkg/controller.v1alpha3/trial/trial_controller.go#L259-L291).

The author decided to use the `if-else` control flow at its maximum
potential: the logic goes as deep as three tabs to the right. We cannot
immediately guess which parts are important and which aren't.

```go
func (r *ReconcileTrial) reconcileJob(instance *trialsv1alpha3.Trial, desiredJob *unstructured.Unstructured) (*unstructured.Unstructured, error) {
    var err error
    logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
    apiVersion := desiredJob.GetAPIVersion()
    kind := desiredJob.GetKind()
    gvk := schema.FromAPIVersionAndKind(apiVersion, kind)

    deployedJob := &unstructured.Unstructured{}
    deployedJob.SetGroupVersionKind(gvk)
    err = r.Get(context.TODO(), types.NamespacedName{Name: desiredJob.GetName(), Namespace: desiredJob.GetNamespace()}, deployedJob)
    if err != nil {
        if errors.IsNotFound(err) {
            if instance.IsCompleted() {
                return nil, nil
            }
            logger.Info("Creating Job", "kind", kind,
                "name", desiredJob.GetName())
            err = r.Create(context.TODO(), desiredJob)
            if err != nil {
                logger.Error(err, "Create job error")
                return nil, err
            }
            eventMsg := fmt.Sprintf("Job %s has been created", desiredJob.GetName())
            r.recorder.Eventf(instance, corev1.EventTypeNormal, JobCreatedReason, eventMsg)
            msg := "Trial is running"
            instance.MarkTrialStatusRunning(TrialRunningReason, msg)
        } else {
            logger.Error(err, "Trial Get error")
            return nil, err
        }
    } else {
        if instance.IsCompleted() && !instance.Spec.RetainRun {
            if err = r.Delete(context.TODO(), desiredJob, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
                logger.Error(err, "Delete job error")
                return nil, err
            } else {
                eventMsg := fmt.Sprintf("Job %s has been deleted", desiredJob.GetName())
                r.recorder.Eventf(instance, corev1.EventTypeNormal, JobDeletedReason, eventMsg)
                return nil, nil
            }
        }
    }

    return deployedJob, nil
}
```

The outline of this function doesn't tell us anything about what the flow
is and where the important logic is. Having deeply nested `if-else`
statements hurt Go's glanceability: where is the "happy path"? Where are
the "error paths"?

<img src="before.png" width="500"/>

By refactoring and removing `else` statements, we obtain a more coherent
aligned-to-the-left path:

<img src="after.png" width="500"/>

The green line represents the "core logic" and is at the minimum
indentation level. The red line represents anything out of the ordinary:
error handling and guards.

And since our eyes are very good at following lines, the _line of sight_
(the green line) guides us and greatly improves the experience of glancing
at a piece of code.

Here is the actual code I rewrote:

```go
func (r *ReconcileTrial) reconcileJob(instance *trialsv1alpha3.Trial, desiredJob *unstructured.Unstructured) (*unstructured.Unstructured, error) {
    var err error
    logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
    apiVersion := desiredJob.GetAPIVersion()
    kind := desiredJob.GetKind()
    gvk := schema.FromAPIVersionAndKind(apiVersion, kind)

    deployedJob := &unstructured.Unstructured{}
    deployedJob.SetGroupVersionKind(gvk)

    err = r.Get(context.TODO(), types.NamespacedName{Name: desiredJob.GetName(), Namespace: desiredJob.GetNamespace()}, deployedJob)
    switch {
    case errors.IsNotFound(err) && instance.IsCompleted():
        // Job deleted and trial completed, nothing left to do.
        return nil, nil
    case errors.IsNotFound(err):
        // Job deleted, we must create a job.
        logger.Info("Creating Job", "kind", kind, "name", desiredJob.GetName())
        err = r.Create(context.TODO(), desiredJob)
        if err != nil {
            logger.Error(err, "Create job error")
            return nil, err
        }
        eventMsg := fmt.Sprintf("Job %s has been created", desiredJob.GetName())
        r.recorder.Eventf(instance, corev1.EventTypeNormal, JobCreatedReason, eventMsg)
        msg := "Trial is running"
        instance.MarkTrialStatusRunning(TrialRunningReason, msg)

        return nil, nil
    case err != nil:
        logger.Error(err, "Trial Get error")
        return nil, err
    }

    if !instance.IsCompleted() || instance.Spec.RetainRun {
        eventMsg := fmt.Sprintf("Job %s has been deleted", desiredJob.GetName())
        r.recorder.Eventf(instance, corev1.EventTypeNormal, JobDeletedReason, eventMsg)
        return nil, nil
    }

    err = r.Delete(context.TODO(), desiredJob, client.PropagationPolicy(metav1.DeletePropagationForeground))
    if err != nil {
        logger.Error(err, "Delete job error")
        return nil, err
    }

    return deployedJob, nil
}
```

You can also take a look at Matt Ryer's _Idiomatic Go Tricks_ where he
presents some ways of keeping your code as readable as possible:

{{< youtube yeetIgNeIkc >}}

---

Wish to comment? Here is the Twitter thread for that post:

{{< twitter 1198259761722122240 >}}
