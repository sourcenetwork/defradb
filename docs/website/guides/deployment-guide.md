---
sidebar_label: Deployment Guide
sidebar_position: 70
---
# A Guide to DefraDB Deployment
DefraDB aspires to be a versatile database, supporting both single-node and clustered deployments. In a clustered setup, multiple nodes collaborate seamlessly. This guide walks you through deploying DefraDB, from single-node configurations to cloud and server environments. Let’s begin.

## Prerequisites
The prerequisites listed in this section should be met before starting the deployment process.

**Pre-Compiled Binaries** - Each release has its own set of pre-compiled binaries for different Operating Systems. Obtain the pre-compiled binaries for your operating system from the [official releases](https://github.com/sourcenetwork/defradb/releases).

### Bare Metal Deployment

For Bare Metal deployments, there are two methods available:

- ### Building from Source

Ensure Git, Go and make are installed for all your development environments.

1. **Unix (Mac and Linux)** - The main thing required is the [Go language toolchain](https://go.dev/dl/), which is supported up to Go 1.20 in DefraDB due to the current dependencies.
2. **Windows** - Install the [MinGW toolchain](https://www.mingw-w64.org/) specific to GCC and add the [Make toolchain](https://www.gnu.org/software/make/).

Follow these steps to build from source:

1. Run git clone to download the [DefraDB repository](https://github.com/sourcenetwork/defradb#install) to your local machine.
2. Navigate to the repository using `cd`.
3. Execute the Make command to build a local DefraDB setup with default configurations.
4. Set the compiler and build tags for the playground: `GOFLAGS="-tags=playground"` 

#### Build Playground

Refer to the Playground Basics Guide for detailed instructions.

1. Compile the playground separately using the command: `make deps:playground`
2. This produces a bundle file in a folder called dist.
3. Set the environment variable using the [NodeJS language toolchain](https://nodejs.org/en/download/current) and npm to build locally on your machine. The JavaScript and Typescript code create an output bundle for the frontend code to work.
4. Build a specific playground version of DefraDB. Use the go flags environment variable, instructing the compiler to include the playground directly embedded in all files. Execute the [go binary embed](https://pkg.go.dev/embed) command, producing a binary of approximately 4MB.



- ### Docker Deployments

Docker deployments are designed for containerized environments. The main prerequisite is that Docker should be installed on your machine.


The steps for Docker deployment are as follows:

1. Install Docker by referring to the [official Docker documentation](https://docs.docker.com/get-docker/).
2. Navigate to the root of the repository where the Dockerfile is located.
3. Run the following command: 
`docker build -t defra -f tools/defradb.containerfile `


**Note**: The period at the end is important and the -f flag specifies the file location.

The container file is in a subfolder called tools: `path: tools/defradb.containerfile`

Docker images streamline the deployment process, requiring fewer dependencies. This produces a DefraDB binary file for manual building and one-click deployments, representing the database in binary form as a system.

## Deployment

### Manual Deployment

DefraDB is a single statically built binary with no third-party dependencies. Similar to bare metal, it can run on any cloud or machine. Execute the following command to start DefraDB:
`defradb start --store badger`



### AWS Environment

For deploying to an AWS environment, note the following:

- Deploy effortlessly with a prebuilt [AMI](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html) (Amazon Machine Image) featuring DefraDB.
- Access the image ID or opt for the convenience of the Amazon Marketplace link.
- Refer to [AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EC2_GetStarted.html) for an easy EC2 instance launch with your specified image size.
- Customize your setup using Packer and Terraform scripts in this directory: `tools/cloud/aws/packer`

 

### Akash Deployments

For detailed instructions on deploying DefraDB with Akash, refer to the [Akash Deployment Guide](https://nasdf-feat-akash-deploy.docs-source-network.pages.dev/guides/akash-deployment).

 

## Configurations

- The default root directory on Unix machines is `$HOME/.defradb`. For Windows it is `%USERPROFILE%\.defradb`​.
- Specifiy the DefraDB folder with this command: `defradb --rootdir <path> start`.
- The default directory for where data is specified is `<rootdir>/data`.

 

## Storage Engine

The storage engines currently used include:

- Fileback persistent storage powered the [Badger](https://github.com/dgraph-io/badger%5D ) database.
- [In-Memory Storage](https://github.com/sourcenetwork/defradb/blob/develop/datastore/memory/memory.go) which is B-Tree based, ideal for testing does not work with the file system. It is specified with this flag: `--store memory`

 

## Network and Connectivity

As a P2P database, DefraDB requires two ports for node communication, they include:

 

1. **API Port**: It powers the HTTP API, handling queries  from the client to the database  and various API commands. The default port number is *9181*.

2. **P2P Port**: It facilitates communication between nodes, supporting data sharing, synchronization, and replication. The default port no is *9171*.

 

The P2P networking functionality can't be disabled entirely, but you can use the `defradb start --no-p2p`​ command through the config files and CLI to deactivate it.

 

### Port Customization

The API port can be specified using the [bind address](https://docs.libp2p.io/concepts/fundamentals/addressing/):

`API: --url <BIND_ADDRESS>:<PORT>`

For P2P use the P2P adder to a multi-address:

`--p2paddr <multiaddress>`

Here is an [infographic](https://images.ctfassets.net/efgoat6bykjh/XQrDLqpkV06rFhT24viJc/1c2c72ddebe609c80fc848bfa9c4771e/multiaddress.png) to further understand multi-address.


## The Peer Key

Secure communication between nodes in DefraDB is established with a unique peer key for each node. Key details include:

 

- The peer key is automatically generated on startup, replacing the key file in a specific path.
- There is no current method for generating a new key except for overwriting an existing one.
- The peer key type uses a specific elliptic curve, called an Ed25519, which can be used to generate private keys.
- In-memory mode generates a new key with each startup.
- The config file located at `<rootdir>/config.yaml` is definable and used for specification.
- Additional methods for users to generate their own Ed25519 key: 
openssl genpkey -algorithm ed25519 -text

## Future Outlook

As DefraDB evolves, the roadmap includes expanding compatibility with diverse deployment environments:

- **Google Cloud Platform (GCP)**: Tailored deployment solutions for seamless integration with GCP environments.
- **Kubernetes**: Optimization for Kubernetes deployments, ensuring scalability and flexibility.
- **Embedded/IoT for Small Environments**: Adaptations to cater to the unique demands of embedded systems and IoT applications.
- **Web Assembly (WASM) Deployments**: Exploring deployment strategies utilizing Web Assembly for enhanced cross-platform compatibility.

 