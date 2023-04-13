// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

variable "ports" {
  description = "ingress ports to open on security group"
  default = [
    {
      from   = 1981
      to     = 1981
      source = "0.0.0.0/0"
    }
  ]
}

variable "region" {
  description = "aws region e.g. us-east-1, us-west-2"
}

variable "organization" {
  description = "organization name"
  default     = "source"
}

variable "environment" {
  description = "environment e.g. dev, qa, prod"
}

variable "deployment_type" {
  description = "deployment type e.g. console|cloudformation|terraform"
}

variable "deployment_repo" {
  description = "deployment repo"
}

variable "ami_name_filter" {
  description = "Filter to use to find the AMI by name"
  default     = "source-defradb-*-*"
}

variable "ami_owner" {
  description = "Filter for the AMI owner"
  default     = "self"
}

variable "instance_type" {
  description = "Type of EC2 instance"
  default     = "t2.micro"
}

variable "keypair" {
  description = "Key pair to access the EC2 instance"
  default     = "default"
}


locals {
  common_tags = {
    organization = var.organization
    environment  = var.environment
  }
}
