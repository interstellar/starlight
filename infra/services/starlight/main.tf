provider "aws" {
  version = "~> 1.22"
  region  = "${var.region}"
}

resource "aws_vpc" "default" {
  cidr_block           = "${var.cidr_block}"
  enable_dns_hostnames = true

  tags {
    Name = "${var.vpc_name}"
  }
}

resource "aws_internet_gateway" "default" {
  vpc_id = "${aws_vpc.default.id}"
}

resource "aws_route" "internet_access" {
  route_table_id         = "${aws_vpc.default.main_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.default.id}"
}

resource "aws_subnet" "a" {
  vpc_id                  = "${aws_vpc.default.id}"
  cidr_block              = "${cidrsubnet(var.cidr_block, 8, 1)}"
  map_public_ip_on_launch = true

  tags {
    Name = "${var.vpc_name}"
  }
}

resource "aws_security_group" "starlight" {
  vpc_id      = "${aws_vpc.default.id}"
  name        = "starlight-node"
  description = "permit 22,80(redirect), and 443 incoming. permit all outgoing"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Public SSH access"
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Public peer access"
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Should redirect to https"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All outgoing traffic allowed"
  }
}

resource "aws_key_pair" "default" {
  key_name   = "${var.vpc_name}-deployer-key"
  public_key = "${var.ssh_public_key}"
}

resource "aws_instance" "default" {
  connection {
    user = "ubuntu"
  }

  ami                    = "${var.ami}"
  instance_type          = "${var.instance_type}"
  vpc_security_group_ids = ["${aws_security_group.starlight.id}"]
  subnet_id              = "${aws_subnet.a.id}"
  key_name               = "${aws_key_pair.default.key_name}"

  root_block_device = {
    volume_size = 100
  }

  tags {
    Name = "${var.vpc_name}-1"
  }

  provisioner "local-exec" "build-starlightd" {
    working_dir = "${var.working_dir}"

    environment {
      GOARCH = "amd64"
      GOOS   = "linux"
    }

    command = "go build -o /tmp/starlightd ./cmd/starlightd"
  }

  provisioner "file" {
    source      = "/tmp/starlightd"
    destination = "/home/ubuntu/starlightd"
  }

  provisioner "file" {
    source      = "./starlight.service"
    destination = "/home/ubuntu/starlight.service"
  }

  provisioner "file" {
    source      = "./starlight.socket"
    destination = "/home/ubuntu/starlight.socket"
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /home/ubuntu/starlightd",
      "sudo mv /home/ubuntu/starlight.service /etc/systemd/system/",
      "sudo mv /home/ubuntu/starlight.socket /etc/systemd/system/",
      "sudo systemctl daemon-reload",
      "sudo systemctl enable starlight.socket",
      "sudo systemctl enable starlight",
      "sudo systemctl start starlight.socket",
    ]
  }
}

resource "aws_eip" "default" {
  instance = "${aws_instance.default.id}"
  vpc      = true
}

resource "aws_route53_zone" "default" {
  name = "${var.domain_name}"
}

resource "aws_route53_record" "default" {
  zone_id = "${aws_route53_zone.default.zone_id}"
  name    = "starlight.${var.domain_name}."
  type    = "A"
  ttl     = "60"
  records = ["${aws_eip.default.public_ip}"]
}
