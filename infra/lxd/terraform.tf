terraform {
  backend "etcdv3" {
    endpoints = [ "localhost:2379" ]
    lock      = true
  }
}
