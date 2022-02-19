# Virtual Garden Landscaper Component for [Gardener](https://gardener.cloud)

🚧 This repository is heavily under construction and should be considered experimental.

This repository contains the implementation to deploy the virtual garden, consisting basically of two etcd's, a
kube apiserver, and kube controller manager.

## What is a "Virtual Garden"?
A "virtual garden" is made up of `garden-etcd-main`, `garden-etcd-events`, `garden-apiserver` and 
`gardener-controller-manager`.   This apiserver and etcd serve as the “data container” for all 
the Gardener resources (Seed, Shoot and so on). With this we have full control over the used k8s 
version and we are independent from the underlying initial cluster.

Specifically, this setup registers your root cluster as a Seed cluster by:
* deploying the gardenlet to the root cluster 
* set the Kubeconfig to the virtual Garden in the `gardenClientConnection`
* Fallback in the `seedClientConnection` to the root cluster

Deploying this way has backup and disaster recovery advantages. Gardener backups are isolated from the
root cluster backup and restoring the root cluster can be completed without API objects of the garden 
landscape being reconciled before the operator is ready for them.

## Usage
There are three options for deployment:

* [Executing a make target](./docs/deploy-virtual-garden-with-make-target.md)

* [Deploying with Landscaper](./docs/deploy-virtual-garden-with-landscaper.md)

* [Local Development](docs/local-development.md)

## [Feedback and Support](https://github.com/gardener/gardener#feedback-and-support)

## [Learn More!](https://github.com/gardener/gardener/blob/master/README.md#learn-more)
