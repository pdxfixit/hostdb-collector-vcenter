# HostDB Collector for vCenter

Queries the vCenter REST API, to get data about all of the virtual machines in each instance, and sends that data to HostDB.

## Getting Started

This section will describe the process of developing the collector.
Please see [Deployment](#deployment) for notes on how the collector is used when deployed.

### Prerequisites

The vCenter collector requires a few things to operate:

* A list (inventory) of the vCenter instances to query.
* A HostDB instance to write to.

For development, you'll also need:

* Docker
* Golang >= v1.11

### Installing

The collector is a golang binary, and after compilation, can be run on any Linux x86 system. No installation necessary.

## Running tests

This should be as simple as `go test`.

## Deployment

The vCenter collector ran from a container in the build system on a regular schedule.

## Built With

Build will run tests, compile the golang binary, create a container including the binary, and upload that container image to the registry.

## Authors & Support

- Email: info@pdxfixit.com
