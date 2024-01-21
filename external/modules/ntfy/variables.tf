variable "auth" {
  type = object({
    url      = string
    username = string
    password = string
  })
}
