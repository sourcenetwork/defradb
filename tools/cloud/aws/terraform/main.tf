// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

data "aws_ami" "ami" {
  most_recent = true
  owners      = ["${var.ami_owner}"]

  filter {
    name   = "name"
    values = ["${var.ami_name_filter}"]
  }
  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_instance" "instance" {
  ami             = data.aws_ami.ami.id
  instance_type   = var.instance_type
  security_groups = ["${aws_security_group.sg.name}"]
  key_name        = var.keypair
  tags            = local.common_tags
}


resource "aws_security_group" "sg" {
  dynamic "ingress" {
    for_each = var.ports[*]
    content {
      from_port        = ingress.value.from
      to_port          = ingress.value.to
      protocol         = "tcp"
      cidr_blocks      = ingress.value.from != 1433 ? [ingress.value.source] : null
      ipv6_cidr_blocks = ingress.value.source == "::/0" ? [ingress.value.source] : null
      security_groups  = ingress.value.from == 1433 ? [ingress.value.source] : null
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.common_tags
}
