<!--[metadata]>
+++
title = "Spotinst Driver for Docker Machine"
description = "Spotinst Driver for Docker Machine"
keywords = ["machine, spotinst, driver"]
[menu.main]
+++
<![end-metadata]-->
# Docker Machine Driver of Spotinst

Create instances on [Spotinst](https://spotinst.com/) using Docker-Machine.

```bash
docker-machine create -d spotinst
```

# Requierments
For Docker-Machine to connect to Spotinst Elastigroup You will need:
 * Spotinst Account
 * Spotinst Token
 * Elastigroup with:
    * Docker-Machine Supported OS AMI (Amazon-Linux isn't supported by Docker-Machine)
    * Security Group with inbound SSH (22) and Docker-Machine (2376) ports open
 * All required parameters from the [Options](#options) section fulfilled
    
 

creates docker instances on Spotinst Elastigroup


## Installation

Please see latest version of the driver with instructions how to install in: [Releases](https://github.com/spotinst/docker-machine-driver-spotinst/releases)

## Options

```bash
docker-machine create -d spotinst --help
```
 Option Name                                          | Description                                           | required
------------------------------------------------------|------------------------------------------------------|----|
``--spotinst-account`` |Spotint Account ID |**yes**|
``--spotinst-elastigroup-id``|ElastGroup ID in the relevant account to fill in servers| **yes** |
``--spotinst-token``|Spotinst Token from you organization| **yes** |
``--spotinst-sshkey-path``|Local path to the pem file of the Elastigroup| **yes** |
``--use-public-ip``|Boolean flag (means do not get any value) that determines if to use public IP or private IP| No |
``--ssh-user``|Username for server SSH connection using the pem| No |

## Examples

The following example creates a server called `dev` on Spotinst Elastigroup 
```apple js
docker-machine create -d spotinst --spotinst-account "act-12345" --spotinst-elastigroup-id "sig-12345" --spotinst-token "<Token>" --spotinst-sshkey-path /home/ubuntu/pems/myssh.pem" --use-public-ip dev
```



