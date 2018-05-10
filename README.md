<!--[metadata]>
+++
title = "Spotinst Driver for Docker Machine"
description = "Spotinst Driver for Docker Machine"
keywords = ["machine, spotinst, driver"]
[menu.main]
+++
<![end-metadata]-->
# Docker Machine Driver of Spotinst

Create machines on [Spotinst](https://spotinst.com/) using Docker-Machine.

```bash
docker-machine create -d spotinst
```

# Requierments
For Docker-Machine to connect to Spotinst ElastiGroup You will need:
 * Spotinst Account
 * Spotinst Token
 * Elastic Group with:
    * Docker-Machine Supported OS AMI (Amazon-Linux does not supported)
    * Security Group with inbound of SSH (22) and Docker-Machine (2376) ports open
 * Fill in the required parameters in from the option section
    
 

creates docker instances on Spotinst ElastiGroup


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
``--spotinst-sshkey-path``|Local path to the pem file of the ElastiGroup| **yes** |
``--use-public-ip``|Boolean flag (means do not get any value) that determines if to use public IP or private IP| No |
``--ssh-user``|Username for server SSH connection using the pem| No |

## Examples

The below example is for creating server call dev using on Spotinst ElastiGroup 
```apple js
docker-machine create -d spotinst --spotinst-account "act-12345" --spotinst-elastigroup-id "sig-12345" --spotinst-token "<Token>" --spotinst-sshkey-path /home/ubuntu/pems/myssh.pem" --use-public-ip dev
```



