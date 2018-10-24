variable "region" {
  type    = "string"
  default = "us-west-2"
}

variable "ssh_public_key" {
  type    = "string"
}

variable "instance_type" {
  type    = "string"
  default = "t3.large"
}

variable "vpc_name" {
  type    = "string"
  default = "starlight"
}

variable "domain_name" {
  type    = "string"
  default = "i10rint.com"
}

variable "working_dir" {
  type = "string"
}

variable "ami" {
  type = "string"

  # I10R's Ubuntu 16
  default = "ami-0dbf15b2329118870"
}

variable "cidr_block" {
  type    = "string"
  default = "172.39.0.0/16"
}
