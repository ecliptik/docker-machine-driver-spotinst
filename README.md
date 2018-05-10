<!--[metadata]>
+++
title = "Spotinst Driver for Docker Machine"
description = "Spotinst Driver for Docker Machine"
keywords = ["machine, spotinst, driver"]
[menu.main]
+++
<![end-metadata]-->
# Docker Machine Driver of Spotinst

Create machines on [Spotinst](https://spotinst.com/).  You will need an Spotinst Account, Spotinst Token and a Elastic Group using ubuntu AMI.

creates docker instances on Spotinst ElastiGroup

```bash
docker-machine create -d spotinst
```

## Installation

Please see latest version of the driver with instructions how to install in: [Releases](https://github.com/spotinst/docker-machine-driver-spotinst/releases)

## Options

```bash
docker-machine create -d aliyunecs --help
```
 Option Name                                          | Description                                           | required
------------------------------------------------------|------------------------------------------------------|----|
``--spotinst-account`` |Spotint Account ID |**yes**|
``--spotinst-elastigroup-id``|ElastGroup ID in the relevant account to fill in servers| **yes** |
``--spotinst-token``|Spotinst Token from you organization| **yes** |
``--spotinst-sshkey-path``|Local path to the pem file of the ElastiGroup| **yes** |
``--use-public-ip``|Boolean flag (means do not get any value) that determines if to use public IP or private IP| No |


