# Oracle Cloud Instance Generator

### Overview

A small tool to automatically create a new instance on Oracle Cloud Infrastructure. 

Oracle's free tier provides up to  4 CPUs and 24Gb of RAM and 200GB of block storage for the ARM tier (and 2 small AMD64 VMs).
However, resources for the free tier on OCI are limited, and new resources can only be created if sufficient host capacity is available.

This may result in the following error message:

```
Out of capacity for shape VM.Standard.A1.Flex in availability domain AD-1. Create the instance in a different availability domain or try again later. If you specified a fault domain, try creating the instance without specifying a fault domain, otherwise try creating the instance in a different availability domain. If that doesn’t work, please try again later. Learn more about host capacity.
```

Unfortunately, it is not possible to create a new instance in all availability zones at the same time, or create it automatically until enough resources are available, which makes this a time consuming task.

This project aims to solve this by attempting to create a new instance in all availability zones after a reasonable time interval.
This allows you to step back from manually creating OCI instances, but rather start this tool and wait until it is finished.

Usually, this project is intended to run over docker or kubernetes, but can also be used locally if you set the environment variables.

The docker image is available at Dockerhub: [mowirth/oci-instance-creator](https://hub.docker.com/r/mowirth/oci-instance-creator)) for ARM and AMD machines.

### Required Configuration

To create an instance, the following information is required:
- Oracle Cloud User ID
- Oracle Cloud Subnet ID
- Oracle Cloud Image ID
- Oracle Cloud Tenancy ID
- Oracle Cloud Region
- Oracle Cloud API Key
- Oracle Cloud API Key Fingerprint

The fastest way to fetch all required values is to go to the instance section, click on Inspect, configure your instance and send it, and review the API request afterwards.
However, using the api key from the request itself will not work, it is mandatory to create a new API key.

Finally, a more detailed tutorial is available here: [https://github.com/hitrov/oci-arm-host-capacity](https://github.com/hitrov/oci-arm-host-capacity)

##### Oracle API Key | Key Fingerprint
We use an API Token generated on the user profile below the API Keys section (click [OCI User API Keys](https://cloud.oracle.com/identity/domains/my-profile/api-keys)).
Afterwards, click on **Add API key**, select *Generate API key pair* and download the private key.

The location of the private key must be specified using the keypath variable and must be readable by the user that executes the tool.
The keypath can be specified with ```KEY_PATH```, the default path is ```$(pwd)/oci.key```.

For Docker/Kubernetes, the path is automatically set to ```/keys/oci.key```. 
In the docker example, the key is expected to be in the same directory as the docker-compose file, in the kubernetes example please update the example or create the secret yourself.

Additionally, the fingerprint of the key must be stored as ```OCI_FINGERPRINT```, which should look similar to this: ```11:22:33:44:ff:ff:00:11:22:33:44:55:66:77:88:99```.

Finally, you will receive the OCI_FINGERPRINT, OCI_USER_ID, OCI_REGION, OCI_TENANCY_ID in the configuration file preview after creating the API Key (make sure to click on Add after downloading the key).

##### Oracle Region

This defines the region that your Oracle Cloud Account is located in. 
It is the same region you did select during registration.

It should look similar to this: ```eu-frankfurt-1```

The region must be stored as ```OCI_REGION``` environment variable.

##### Oracle Cloud User ID

The User ID is also located on the user profile of OCI, at the OCID field.

It must be configured as ```OCI_USER_ID```, and should look similar to this: ```ocid1.user.oc1..verlongrandomstring```

Jump to [OCI User Page](https://cloud.oracle.com/identity/domains/my-profile).

##### Oracle Cloud Tenancy ID

Your tenancy ID can be reviewed at the Tenancy details in the administration section.

It must be configured with ```OCI_TENANCY_ID``` and should look similar to this: ```ocid1.tenancy.oc1..verylongrandomstring```

Jump to [OCI Tenancy Details](https://cloud.oracle.com/tenancy).

---

##### Oracle Cloud Image ID 

There is unfortunately not a nice way to get the image ID yet. To obtain it, either create an instance with it, and then view the instance details, click on the link at *Image:* and use the OCID.
The Image ID is the same for image version/region. Alternatively, inspect the API request and copy the image id from there.

Your image ID must be configured with ```OCI_IMAGE_ID``` and should look similar to this: ```ocid1.image.oc1.eu-frankfurt-1.verylongrandomstring```

##### Oracle Cloud Subnet ID

```OCI_SUBNET_ID``` configures subnet that the cloud should be attached to must be created prior to creating this subnet.

To create the subnet:
- Go to Oracle Cloud --> Virtual cloud networks
- Create a new network or select an existing network
- On the main tab, select the appropriate subnet in the list
- On the subnet info page, the subnet id is located at OCID. It should look similar to ```ocid1.subnet.oc1.your-region.verylongrandomstring```

Jump to [OCI Virtual Cloud Networks](https://cloud.oracle.com/networking/vcns)

##### SSH Key

In order to access your instance, you must provide an SSH key during generation of your instance.
This project does not support generating your keys during the creation process. 

To Provide your SSH key, please use the ```SSH_KEY``` environment variable.


### Additional (optional) configuration

There are additional configuration methods in case you wish to create a different instance.

| Env Variable            | Description                                                                                 | Default                  |
|-------------------------|---------------------------------------------------------------------------------------------|--------------------------|
| DISPLAY_NAME            | Set the display name of the                                                                 | Unix Timestamp as string |
| SHAPE                   | Configure the shape. Default is the ARM machine                                             | VM.Standard.A1.Flex      |
| CPUS                    | Count of CPUs. Is used to determine RAM (4 CPUS/24GB)                                       | 4                        |
| VOLUME_SIZE             | Size of the boot volume. Must be at least 50Gb. 200Gb is max for free tier                  | 50                       |
| CREATE_INTERVAL_SECONDS | Seconds until next run to create an instance in all availability zones. Default is 1 minute | 60                       |
| CREATE_ZONE_SECONDS     | Seconds to wait between each create attempt at an availability zone. Default is 10 seconds  | 10                       |

### Kubernetes

You can natively run this project on Kubernetes, either as a job or as an init-container etc. 

The job file expects a standard config map and the key file. 

To create the secret, add it from file:

```kubectl create secret generic oci-secret --from-file=oci.key=./oci.key```

Example files are provided in the k8s folder.


### Contributions / Thanks

Contributions/Questions are always welcome, feel free to open an issue for further questions.

Thanks to https://github.com/hitrov/oci-arm-host-capacity for providing the initial approach and most of the configuration required for OCI.