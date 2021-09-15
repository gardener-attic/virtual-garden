# Deploying the Virtual Garden with Make Target

There are make targets to deploy and undeploy the virtual garden.  

### Configuration

Before you can use the make targets below, you must adjust the configuration in the 
[imports.yaml](../example/imports.yaml) file. 

Provide at least the kubeconfig of the host cluster on which the virtual garden should be deployed. 
Insert the kubeconfig at `.cluster.spec.config.kubeconfig`.

### Deploy

You can deploy the virtual garden with the command

```shell script
make start
```

The virtual garden is deployed to the namespace specified at `.hostingCluster.namespace` in the 
[imports.yaml](../example/imports.yaml) file.

The deploy program writes exports parameters to the [exports.yaml](../example/exports.yaml) file.
It contains in particular a kubeconfig to access the virtual garden in field `.kubeconfigYaml`.

### Undeploy

You can undeploy the virtual garden with the command

```shell script
make delete
```
