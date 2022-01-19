# Deploying the Virtual Garden with [Landscaper](https://github.com/gardener/landscaper)

The virtual garden can be deployed using the 
[container deployer](https://github.com/gardener/landscaper/blob/master/docs/deployer/container.md) of Landscaper.

This requires a cluster on which the landscaper is installed.

### Creating the Container Image

The container deployer executes an image with the deploy logic that is implemented in this project.
The following steps describe how to build this image and push it into an OCI registry.

- Adjust the `REGISTRY` variable in the [Makefile](../Makefile) so that it points to your OCI registry.

- Execute the following command to build the image
  ```shell script
  make docker-images
  ``` 
  
- Login to the OCI registry and execute the following command to push the image into the OCI registry.
  ```shell script
  make docker-push
  ``` 

### Creating the Component Descriptor

The following command creates a component descriptor for the virtual garden component and pushes it into the
OCI registry.

```shell script
make cnudie
``` 

The component descriptor contains the list of all resources required for the deployment of the virtual garden: 

- the blueprint,  
- the image from the previous step, which will be executed by the container deployer,  
- the images of etcd, kube-apiserver, etc. which will be deployed to the runtime cluster of the virtual garden.  

### Creating a Target and an Installation

In order to trigger the deployment of the virtual garden, we must create a Target and an Installation on the landscaper 
cluster. Both must be in the same namespace.

- Create a [Target](https://github.com/gardener/landscapercli/blob/master/docs/examples/target.yaml) custom resource 
on the landscaper cluster. The Target must have the type `landscaper.gardener.cloud/kubernetes-cluster` and contain the 
kubeconfig of the host cluster of the virtual garden.

- Use the following command to create the yaml manifest of an Installation:

  ```shell script
  make create-installation
  ```

- Adjust the Installation:

  - Field `.spec.imports.targets[0].target` must contain a hash sign `#` followed by the name of the Target created in 
    the previous step.

  - Field `.spec.importDataMappings` contains the configuration of the virtual garden.

- Apply the Installation to the landscaper cluster.
